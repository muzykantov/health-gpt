package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/muzykantov/health-gpt/chat"
	"golang.org/x/net/proxy"
)

var (
	ErrDeepSeekUnsupportedRole        = errors.New("unsupported role")
	ErrDeepSeekUnsupportedContentType = errors.New("unsupported content type")
	ErrDeepSeekRequestFailed          = errors.New("request failed")
)

// DeepSeekOption defines a configuration function for the client.
type DeepSeekOption func(*DeepSeek)

// DeepSeek implements a client for interacting with DeepSeek API.
type DeepSeek struct {
	client *http.Client
	// Configuration parameters.
	model       string  // Model to use.
	temperature float64 // Generation temperature (0.0-2.0).
	topP        float64 // Top-p sampling (0.0-1.0).
	maxTokens   int64   // Maximum number of tokens in response.
	socksProxy  string  // SOCKS proxy address.
	baseURL     string  // Base API URL.
	apiKey      string  // API key for authorization.
}

// DeepSeekWithModel sets the model to use.
func DeepSeekWithModel(model string) DeepSeekOption {
	return func(c *DeepSeek) {
		if model != "" {
			c.model = model
		}
	}
}

// DeepSeekWithTemperature sets the generation temperature.
func DeepSeekWithTemperature(temperature float64) DeepSeekOption {
	return func(c *DeepSeek) {
		if temperature != 0 {
			c.temperature = temperature
		}
	}
}

// DeepSeekWithTopP sets the top-p sampling parameter.
func DeepSeekWithTopP(topP float64) DeepSeekOption {
	return func(c *DeepSeek) {
		if topP != 0 {
			c.topP = topP
		}
	}
}

// DeepSeekWithMaxTokens sets the maximum number of tokens.
func DeepSeekWithMaxTokens(maxTokens int64) DeepSeekOption {
	return func(c *DeepSeek) {
		if maxTokens != 0 {
			c.maxTokens = maxTokens
		}
	}
}

// DeepSeekWithSocksProxy sets the SOCKS proxy.
func DeepSeekWithSocksProxy(socksProxy string) DeepSeekOption {
	return func(c *DeepSeek) {
		if socksProxy != "" {
			c.socksProxy = socksProxy
		}
	}
}

// DeepSeekWithBaseURL sets the base API URL.
func DeepSeekWithBaseURL(baseURL string) DeepSeekOption {
	return func(c *DeepSeek) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// NewDeepSeek creates a new DeepSeek instance with the given options.
func NewDeepSeek(apiKey string, opts ...DeepSeekOption) (*DeepSeek, error) {
	c := &DeepSeek{
		model:       "deepseek-chat",
		temperature: 0.1,
		topP:        1.0,
		maxTokens:   2048,
		baseURL:     "https://api.deepseek.com",
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

// DeepSeekMessage represents a message in DeepSeek API format.
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest represents a request to DeepSeek API.
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Stream      bool              `json:"stream"`
	Temperature float64           `json:"temperature"`
	TopP        float64           `json:"top_p"`
	MaxTokens   int64             `json:"max_tokens"`
}

// DeepSeekResponse represents a response from DeepSeek API.
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// CompleteChat implements the Completion interface.
func (c *DeepSeek) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	deepseekMessages := make([]DeepSeekMessage, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrDeepSeekUnsupportedContentType,
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
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrDeepSeekUnsupportedRole,
				msg.Sender,
			)
		}

		deepseekMessages[i] = DeepSeekMessage{
			Role:    role,
			Content: content,
		}
	}

	// Create request
	request := DeepSeekRequest{
		Model:       c.model,
		Messages:    deepseekMessages,
		Stream:      false,
		Temperature: c.temperature,
		TopP:        c.topP,
		MaxTokens:   c.maxTokens,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf("%w: %v", ErrDeepSeekRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
				ErrDeepSeekRequestFailed,
				resp.StatusCode,
			)
		}

		return chat.EmptyMessage, fmt.Errorf(
			"%w: %s (%s)",
			ErrDeepSeekRequestFailed,
			errorResponse.Error.Message,
			errorResponse.Error.Type,
		)
	}

	var response DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return chat.EmptyMessage, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrDeepSeekRequestFailed,
		)
	}

	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: response.Choices[0].Message.Content,
	}, nil
}
