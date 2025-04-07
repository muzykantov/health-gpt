package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/metrics"
	"golang.org/x/net/proxy"
)

var (
	ErrMistralUnsupportedRole        = errors.New("unsupported role")
	ErrMistralUnsupportedContentType = errors.New("unsupported content type")
	ErrMistralRequestFailed          = errors.New("request failed")
)

// MistralOption defines a configuration function for the client.
type MistralOption func(*Mistral)

// Mistral implements a client for interacting with Mistral API.
type Mistral struct {
	client *http.Client

	// Configuration parameters.
	model       string  // Model to use.
	temperature float64 // Generation temperature (0.0-1.5).
	topP        float64 // Top-p sampling (0.0-1.0).
	maxTokens   int64   // Maximum number of tokens in response.
	socksProxy  string  // SOCKS proxy address.
	baseURL     string  // Base API URL.
	apiKey      string  // API key for authorization.
}

// MistralWithModel sets the model to use.
func MistralWithModel(model string) MistralOption {
	return func(c *Mistral) {
		if model != "" {
			c.model = model
		}
	}
}

// MistralWithTemperature sets the generation temperature.
func MistralWithTemperature(temperature float64) MistralOption {
	return func(c *Mistral) {
		if temperature != 0 {
			c.temperature = temperature
		}
	}
}

// MistralWithTopP sets the top-p sampling parameter.
func MistralWithTopP(topP float64) MistralOption {
	return func(c *Mistral) {
		if topP != 0 {
			c.topP = topP
		}
	}
}

// MistralWithMaxTokens sets the maximum number of tokens.
func MistralWithMaxTokens(maxTokens int64) MistralOption {
	return func(c *Mistral) {
		if maxTokens != 0 {
			c.maxTokens = maxTokens
		}
	}
}

// MistralWithSocksProxy sets the SOCKS proxy.
func MistralWithSocksProxy(socksProxy string) MistralOption {
	return func(c *Mistral) {
		if socksProxy != "" {
			c.socksProxy = socksProxy
		}
	}
}

// MistralWithBaseURL sets the base API URL.
func MistralWithBaseURL(baseURL string) MistralOption {
	return func(c *Mistral) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// NewMistral creates a new Mistral instance with the given options.
func NewMistral(apiKey string, opts ...MistralOption) (*Mistral, error) {
	c := &Mistral{
		model:       "mistral-small-latest",
		temperature: 0.1,
		topP:        1.0,
		maxTokens:   1024,
		baseURL:     "https://api.mistral.ai",
		apiKey:      apiKey,
		client:      &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.socksProxy != "" {
		dialer, err := proxy.SOCKS5("tcp", c.socksProxy, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}

		transport := &http.Transport{
			DialContext:       dialContext,
			DisableKeepAlives: true,
		}

		c.client = &http.Client{Transport: transport}
	}

	return c, nil
}

// MistralMessage represents a message in Mistral API format.
type MistralMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MistralRequest represents a request to Mistral API.
type MistralRequest struct {
	Model       string           `json:"model"`
	Messages    []MistralMessage `json:"messages"`
	Stream      bool             `json:"stream"`
	Temperature float64          `json:"temperature"`
	TopP        float64          `json:"top_p"`
	MaxTokens   int64            `json:"max_tokens,omitempty"`
}

// MistralResponse represents a response from Mistral API.
type MistralResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ModelName returns LLM's model name.
func (c *Mistral) ModelName() string {
	return fmt.Sprintf("mistral_%s", c.model)
}

// CompleteChat implements the Completion interface.
func (c *Mistral) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	var (
		start    = time.Now()
		provider = "mistral"
		status   = "success"
	)

	mistralMessages := make([]MistralMessage, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			status = "type_error"
			metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrMistralUnsupportedContentType,
				msg.Content,
			)
		}

		var role string
		switch msg.Sender {
		case chat.RoleSystem:
			role = "system"
		case chat.RoleUser:
			role = "user"
		case chat.RoleAssistant:
			role = "assistant"
		default:
			status = "role_error"
			metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrMistralUnsupportedRole,
				msg.Sender,
			)
		}

		mistralMessages[i] = MistralMessage{
			Role:    role,
			Content: content,
		}
	}

	// Create request
	request := MistralRequest{
		Model:       c.model,
		Messages:    mistralMessages,
		Stream:      false,
		Temperature: c.temperature,
		TopP:        c.topP,
	}

	// Only set max_tokens if it's not 0 to avoid sending "max_tokens":0 in the JSON
	if c.maxTokens > 0 {
		request.MaxTokens = c.maxTokens
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		status = "marshal_error"
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
		return chat.EmptyMessage, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/v1/chat/completions", c.baseURL),
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		status = "request_create_error"
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
		return chat.EmptyMessage, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		status = "network_error"
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
		return chat.EmptyMessage, fmt.Errorf("%w: %v", ErrMistralRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status = fmt.Sprintf("http_%d", resp.StatusCode)
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))

		var errorResponse struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: status code %d",
				ErrMistralRequestFailed,
				resp.StatusCode,
			)
		}

		return chat.EmptyMessage, fmt.Errorf(
			"%w: %s (%s)",
			ErrMistralRequestFailed,
			errorResponse.Error.Message,
			errorResponse.Error.Type,
		)
	}

	var response MistralResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		status = "decode_error"
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
		return chat.EmptyMessage, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		status = "empty_response"
		metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrMistralRequestFailed,
		)
	}

	metrics.ObserveRequestDuration(provider, c.model, status, time.Since(start))
	metrics.AddTokens(provider, c.model, response.Usage.PromptTokens, response.Usage.CompletionTokens)

	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: response.Choices[0].Message.Content,
	}, nil
}
