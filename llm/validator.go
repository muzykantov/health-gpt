package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/metrics"
)

// validationPrompt is the system prompt used for validation.
const validationPrompt = `You are a language model output validator.
IMPORTANT: Read the entire SYSTEM PROMPT, CHAT HISTORY, and MODEL RESPONSE carefully before evaluation. Do not make snap judgments. Consider the following aspects:
1. PROMPT COMPLIANCE - the response MUST follow the system prompt structure
2. ACCURACY - factual correctness of information
3. RELEVANCE - appropriate to the user's question
4. SAFETY - contains no harmful content
BE CAREFUL NOT TO REJECT VALID RESPONSES. If the answer is factually correct and follows the system prompt structure, even with minor deviations, consider it valid.
Take your time to thoroughly analyze each aspect before making a judgment.
Return ONLY a RAW JSON with this EXACT structure WITHOUT ANY MARKDOWN OR COMMENTS:
{
"can_send_to_user": true/false,
"follows_prompt": true/false,
"reliability_score": 0.00-1.00,
"reason": "explanation in Russian if can_send_to_user is false or follows_prompt is false"
}
IMPORTANT: Return ONLY valid RAW JSON. NOTHING ELSE.`

// correctionPrompt is the template for requesting corrections.
const correctionPrompt = `[VALIDATOR]. Response requires correction.
REASON: %s
ISSUES:
- Structure compliance: %v
- Can be sent to user: %v
- Reliability score: %.2f
Fix these ISSUES in your response according to the REASON provided. Follow the system prompt structure.
DO NOT ENGAGE IN DIALOG WITH THE VALIDATOR, PROVIDE A NEW CORRECTED RESPONSE.
RELIABILITY SCORE SHOULD AIM FOR 1.00`

const defaultMaxRetryAttempts = 5

// ErrInvalidValidation indicates that the validation response couldn't be processed.
var ErrInvalidValidation = errors.New("invalid validation response")

// ValidationResult represents the structure of the validation response.
type ValidationResult struct {
	CanSendToUser    bool    `json:"can_send_to_user"`
	FollowsPrompt    bool    `json:"follows_prompt"`
	ReliabilityScore float64 `json:"reliability_score"`
	Reason           string  `json:"reason,omitempty"`
}

// ChatCompleter generates responses using a language model.
type ChatCompleter interface {
	ModelName() string
	CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error)
}

// Validator validates language model responses with additional prompts and
// retries if hallucination is suspected.
type Validator struct {
	model     ChatCompleter // Original model
	validator ChatCompleter // Validator model
	maxRetry  int           // Maximum number of retry attempts
	debug     bool          // Append validation results to all responses
	logger    *log.Logger   // Logger for validation operations

	// Model information for metrics
	modelProvider string // Provider of the model (e.g., anthropic, openai)
	modelName     string // Name of the model (e.g., claude-3-sonnet)

	// Validator model information
	validatorProvider string // Provider of the validator model
	validatorName     string // Name of the validator model
}

// NewValidator returns an initialized validator with required model info for metrics.
func NewValidator(
	model ChatCompleter,
	validator ChatCompleter,
	modelProvider string,
	modelName string,
	validatorProvider string,
	validatorName string,
	maxRetry int,
	debug bool,
	logger *log.Logger,
) *Validator {
	if maxRetry <= 0 {
		maxRetry = defaultMaxRetryAttempts
	}

	if logger == nil {
		logger = log.New(log.Writer(), "validator: ", log.LstdFlags)
	}

	return &Validator{
		model:             model,
		validator:         validator,
		maxRetry:          maxRetry,
		debug:             debug,
		logger:            logger,
		modelProvider:     modelProvider,
		modelName:         modelName,
		validatorProvider: validatorProvider,
		validatorName:     validatorName,
	}
}

// ModelName returns name of the original model.
func (v *Validator) ModelName() string {
	return v.model.ModelName()
}

// CompleteChat requests a response from LLM and validates the result.
func (v *Validator) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	var (
		start          = time.Now()
		correctionMsgs = make([]chat.Message, len(msgs))
		retryCount     = 0
		response       chat.Message
		err            error
	)

	copy(correctionMsgs, msgs)

	// Get initial model response
	response, err = v.model.CompleteChat(ctx, correctionMsgs)
	if err != nil {
		v.logger.Printf("[validator] Error getting model response: %v", err)

		metrics.ObserveValidationDuration(v.modelProvider, v.modelName, "initial_error", time.Since(start))

		return chat.EmptyMessage, err
	}

	correctionMsgs = append(correctionMsgs, chat.Message{
		Sender:  chat.RoleAssistant,
		Content: response.Content,
	})

	// Validation and correction loop
	for attempt := range v.maxRetry {
		content, ok := response.Content.(string)
		if !ok {
			v.logger.Printf("[validator] Unexpected response type: %T", response.Content)
			return response, nil
		}

		if v.debug {
			v.logger.Printf("[validator] Model response: %s", content)
		}

		// Validate the response
		valid, err := v.validateResponse(ctx, msgs, response)
		if err != nil {
			v.logger.Printf("[validator] Validation error: %v", err)
			valid = &ValidationResult{
				CanSendToUser:    false,
				FollowsPrompt:    false,
				ReliabilityScore: 0.0,
				Reason:           "Ошибка проверки ответа.",
			}
		}

		v.logger.Printf("[validator] Result: send=%v, follows=%v, score=%.2f",
			valid.CanSendToUser, valid.FollowsPrompt, valid.ReliabilityScore)

		metrics.ObserveValidationScore(v.modelProvider, v.modelName, valid.ReliabilityScore)

		isJSON := json.Valid([]byte(content))

		// If response is valid, return it
		if valid.CanSendToUser && valid.FollowsPrompt {
			if !isJSON && v.debug {
				notification := fmt.Sprintf("\n\n[Проверка ответа пройдена. Достоверность: %.0f%%]",
					valid.ReliabilityScore*100)
				response.Content = content + notification
			}

			metrics.ObserveValidationRetries(v.modelProvider, v.modelName, retryCount)
			metrics.ObserveValidationDuration(v.modelProvider, v.modelName, "success", time.Since(start))

			return response, nil
		}

		// Last attempt - return with warning
		if attempt == v.maxRetry-1 {
			if !isJSON {
				warning := fmt.Sprintf("\n\n[ПРЕДУПРЕЖДЕНИЕ: Этот ответ может быть неточным. Надёжность: %.0f%%. Причина: %s]",
					valid.ReliabilityScore*100, valid.Reason)
				response.Content = content + warning
			}

			metrics.ObserveValidationRetries(v.modelProvider, v.modelName, retryCount)
			metrics.ObserveValidationDuration(v.modelProvider, v.modelName, "max_retries", time.Since(start))

			return response, nil
		}

		correctionMsgs = append(correctionMsgs, chat.Message{
			Sender: chat.RoleUser,
			Content: fmt.Sprintf(
				correctionPrompt,
				valid.Reason,
				valid.FollowsPrompt,
				valid.CanSendToUser,
				valid.ReliabilityScore),
		})

		v.logger.Printf("[validator] Requesting correction, reason: %s", valid.Reason)
		retryCount++

		// Get corrected response for next iteration
		response, err = v.model.CompleteChat(ctx, correctionMsgs)
		if err != nil {
			v.logger.Printf("[validator] Error getting corrected response: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Add response to the history
		correctionMsgs = append(correctionMsgs, chat.Message{
			Sender:  chat.RoleAssistant,
			Content: response.Content,
		})
	}

	return response, nil
}

// validateResponse validates a model response.
func (v *Validator) validateResponse(ctx context.Context, originalMsgs []chat.Message, response chat.Message) (*ValidationResult, error) {
	// Find system prompt in original messages
	var systemPrompt string
	for _, msg := range originalMsgs {
		if msg.Sender == chat.RoleSystem {
			if content, ok := msg.Content.(string); ok {
				systemPrompt = content
				break
			}
		}
	}

	// Create validation messages - we send the entire chat history for context
	validationMsgs := []chat.Message{
		{
			Sender:  chat.RoleSystem,
			Content: validationPrompt,
		},
		{
			Sender: chat.RoleUser,
			Content: fmt.Sprintf("SYSTEM PROMPT: %s\n\nCHAT HISTORY: %s\n\nMODEL RESPONSE: %s",
				systemPrompt,
				formatChatHistory(originalMsgs),
				response.Content,
			),
		},
	}

	// Get validation from the validator model
	validationResponse, err := v.validator.CompleteChat(ctx, validationMsgs)
	if err != nil {
		v.logger.Printf("[validator] Validator request failed: %v", err)
		return nil, err
	}

	// Parse validation result
	content, ok := validationResponse.Content.(string)
	if !ok {
		v.logger.Printf("[validator] Unexpected validation response type: %T", validationResponse.Content)
		return nil, ErrInvalidValidation
	}

	var result ValidationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		v.logger.Printf("[validator] JSON parsing error: %v, response: %s", err, content)
		// If JSON parsing fails, consider it an invalid response
		return &ValidationResult{
			CanSendToUser:    false,
			FollowsPrompt:    false,
			ReliabilityScore: 0.0,
			Reason:           "Не удалось выполнить проверку ответа.",
		}, nil
	}

	return &result, nil
}

// formatChatHistory converts chat messages to a readable format for validation
func formatChatHistory(msgs []chat.Message) string {
	var sb strings.Builder

	for i, msg := range msgs {
		// Skip system messages
		if msg.Sender == chat.RoleSystem {
			continue
		}

		content, ok := msg.Content.(string)
		if !ok {
			continue
		}

		role := "Unknown"
		switch msg.Sender {
		case chat.RoleUser:
			role = "User"
		case chat.RoleAssistant:
			role = "Assistant"
		}

		sb.WriteString(fmt.Sprintf("[%d] %s: %s\n\n", i+1, role, content))
	}

	return sb.String()
}
