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
)

// ValidationPrompt is the system prompt used for validation.
const ValidationPrompt = `You are a language model output validator.
Analyze the response based on:
1. PROMPT COMPLIANCE - the response MUST follow the system prompt structure
2. ACCURACY - factual correctness of information
3. RELEVANCE - appropriate to the user's question
4. SAFETY - contains no harmful content
Return ONLY a RAW JSON with this EXACT structure WITHOUT ANY MARKDOWN OR COMMENTS:
{
"can_send_to_user": true/false,
"follows_prompt": true/false,
"reliability_score": 0.0-1.0,
"reason": "explanation in Russian if can_send_to_user is false or follows_prompt is false"
}
IMPORTANT: Return ONLY valid RAW JSON. NOTHING ELSE.`

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
}

// NewValidator returns an initialized validator.
func NewValidator(
	model,
	validator ChatCompleter,
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
		model:     model,
		validator: validator,
		maxRetry:  maxRetry,
		debug:     debug,
		logger:    logger,
	}
}

// CompleteChat requests a response from LLM and validates the result.
func (v *Validator) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	var (
		response       chat.Message
		err            error
		correctionMsgs = make([]chat.Message, len(msgs))
	)

	copy(correctionMsgs, msgs)

	// Get initial model response
	response, err = v.model.CompleteChat(ctx, correctionMsgs)
	if err != nil {
		v.logger.Printf("[validator] Error getting model response: %v", err)
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

		isJSON := json.Valid([]byte(content))

		// If response is valid, return it
		if valid.CanSendToUser && valid.FollowsPrompt {
			if !isJSON && v.debug {
				notification := fmt.Sprintf("\n\n[Проверка ответа пройдена. Достоверность: %.0f%%]",
					valid.ReliabilityScore*100)
				response.Content = content + notification
			}
			return response, nil
		}

		// Last attempt - return with warning
		if attempt == v.maxRetry-1 {
			if !isJSON {
				warning := fmt.Sprintf("\n\n[ПРЕДУПРЕЖДЕНИЕ: Этот ответ может быть неточным. Надёжность: %.0f%%. Причина: %s]",
					valid.ReliabilityScore*100, valid.Reason)
				response.Content = content + warning
			}
			return response, nil
		}

		correctionMsgs = append(correctionMsgs, chat.Message{
			Sender: chat.RoleUser,
			Content: fmt.Sprintf(
				"[VALIDATOR]. Ответ требует корректировки.\nПРИЧИНА: %s\nПРОБЛЕМЫ:\n"+
					"- Соответствие структуре: %v\n"+
					"- Можно отправить пользователю: %v\n"+
					"- Надежность: %.0f%%\n"+
					"Исправь с учетом указанных проблем. Следуй структуре системного промпта.\n"+
					"НЕ ВЕДИ ДИАЛОГ С VALIDATOR, ОТВЕТЬ С УЧЕТОМ КОРРЕКТИРОВОК.\nНАДЕЖНОСТЬ ДОЛЖНА СТРЕМИТЬСЯ К 100%%",
				valid.Reason,
				valid.FollowsPrompt,
				valid.CanSendToUser,
				valid.ReliabilityScore*100),
		})

		v.logger.Printf("[validator] Requesting correction, reason: %s", valid.Reason)

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
			Content: ValidationPrompt,
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
		return nil, err
	}

	// Parse validation result
	content, ok := validationResponse.Content.(string)
	if !ok {
		return nil, ErrInvalidValidation
	}

	var result ValidationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		v.logger.Printf("[validator] JSON parsing error: %v", err)
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
