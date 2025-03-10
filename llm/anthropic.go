package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/muzykantov/health-gpt/chat"
	"golang.org/x/net/proxy"
)

var (
	ErrAnthropicUnsupportedRole        = errors.New("unsupported role")
	ErrAnthropicUnsupportedContentType = errors.New("unsupported content type")
	ErrAnthropicRequestFailed          = errors.New("request failed")
)

// AnthropicOption defines a configuration function for the client.
type AnthropicOption func(*Anthropic)

// Anthropic implements a client for interacting with Anthropic.
type Anthropic struct {
	client *anthropic.Client
	// Configuration parameters.
	model       string  // Model to use.
	temperature float64 // Generation temperature (0.0-1.0).
	topP        float64 // Top-p sampling (0.0-1.0).
	maxTokens   int64   // Maximum number of tokens in response.
	socksProxy  string  // SOCKS proxy address.
	baseURL     string  // Base API URL.
}

// AnthropicWithModel sets the model to use.
func AnthropicWithModel(model string) AnthropicOption {
	return func(c *Anthropic) {
		if model != "" {
			c.model = model
		}
	}
}

// AnthropicWithTemperature sets the generation temperature.
func AnthropicWithTemperature(temperature float64) AnthropicOption {
	return func(c *Anthropic) {
		if temperature != 0 {
			c.temperature = temperature
		}
	}
}

// AnthropicWithTopP sets the top-p sampling parameter.
func AnthropicWithTopP(topP float64) AnthropicOption {
	return func(c *Anthropic) {
		if topP != 0 {
			c.topP = topP
		}
	}
}

// AnthropicWithMaxTokens sets the maximum number of tokens.
func AnthropicWithMaxTokens(maxTokens int64) AnthropicOption {
	return func(c *Anthropic) {
		if maxTokens != 0 {
			c.maxTokens = maxTokens
		}
	}
}

// AnthropicWithSocksProxy sets the SOCKS proxy.
func AnthropicWithSocksProxy(socksProxy string) AnthropicOption {
	return func(c *Anthropic) {
		if socksProxy != "" {
			c.socksProxy = socksProxy
		}
	}
}

// AnthropicWithBaseURL sets the base API URL.
func AnthropicWithBaseURL(baseURL string) AnthropicOption {
	return func(c *Anthropic) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// NewAnthropic creates a new Anthropic instance with the given options.
func NewAnthropic(apiKey string, opts ...AnthropicOption) (*Anthropic, error) {
	c := &Anthropic{
		model:       string(anthropic.ModelClaude3_7SonnetLatest),
		temperature: 0.1,
		maxTokens:   1024,
	}

	for _, opt := range opts {
		opt(c)
	}

	requestOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if c.baseURL != "" {
		requestOpts = append(requestOpts, option.WithBaseURL(c.baseURL))
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

		requestOpts = append(requestOpts, option.WithHTTPClient(&http.Client{Transport: transport}))
	}

	c.client = anthropic.NewClient(requestOpts...)

	return c, nil
}

// CompleteChat implements the Completion interface.
func (c *Anthropic) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	anthropicMessages := make([]anthropic.MessageParam, 0, len(msgs))

	var systemContent string

	// Extract system message if present
	for _, msg := range msgs {
		if msg.Sender == chat.RoleSystem {
			content, ok := msg.Content.(string)
			if !ok {
				return chat.EmptyMessage, fmt.Errorf(
					"%w: %T",
					ErrAnthropicUnsupportedContentType,
					msg.Content,
				)
			}
			systemContent = content
			break
		}
	}

	// Convert other messages to Anthropic format
	for _, msg := range msgs {
		// Skip system messages as they are handled separately
		if msg.Sender == chat.RoleSystem {
			continue
		}

		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrAnthropicUnsupportedContentType,
				msg.Content,
			)
		}

		switch msg.Sender {
		case chat.RoleUser:
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(
				anthropic.NewTextBlock(content),
			))
		case chat.RoleAssistant:
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(content),
			))
		default:
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrAnthropicUnsupportedRole,
				msg.Sender,
			)
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(c.maxTokens),
		Messages:  anthropic.F(anthropicMessages),
	}

	if c.temperature > 0 {
		params.Temperature = anthropic.F(c.temperature)
	}

	if c.topP > 0 {
		params.TopP = anthropic.F(c.topP)
	}

	if systemContent != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemContent),
		})
	}

	response, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: anthropic request failed: %w",
			ErrAnthropicRequestFailed,
			err,
		)
	}

	if len(response.Content) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no content in response",
			ErrAnthropicRequestFailed,
		)
	}

	// Extract text content from response
	var textContent strings.Builder
	for _, block := range response.Content {
		if block.Text != "" {
			textContent.WriteString(block.Text)
		}
	}

	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: textContent.String(),
	}, nil
}
