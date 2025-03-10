package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"golang.org/x/net/proxy"
)

var (
	ErrOpenAIUnsupportedRole        = errors.New("unsupported role")
	ErrOpenAIUnsupportedContentType = errors.New("unsupported content type")
	ErrOpenAIRequestFailed          = errors.New("request failed")
)

// OpenAIOption defines a configuration function for the client.
type OpenAIOption func(*OpenAI)

// OpenAI implements a client for interacting with OpenAI.
type OpenAI struct {
	client *openai.Client
	// Configuration parameters.
	model       openai.ChatModel // Model to use.
	temperature float64          // Generation temperature (0.0-2.0).
	topP        float64          // Top-p sampling (0.0-1.0).
	maxTokens   int64            // Maximum number of tokens in response.
	socksProxy  string           // SOCKS proxy address.
	baseURL     string           // Base API URL.
}

// OpenAIWithModel sets the model to use.
func OpenAIWithModel(model string) OpenAIOption {
	return func(c *OpenAI) {
		if model != "" {
			c.model = model
		}
	}
}

// OpenAIWithTemperature sets the generation temperature.
func OpenAIWithTemperature(temperature float64) OpenAIOption {
	return func(c *OpenAI) {
		if temperature != 0 {
			c.temperature = temperature
		}
	}
}

// OpenAIWithTopP sets the top-p sampling parameter.
func OpenAIWithTopP(topP float64) OpenAIOption {
	return func(c *OpenAI) {
		if topP != 0 {
			c.topP = topP
		}
	}
}

// OpenAIWithMaxTokens sets the maximum number of tokens.
func OpenAIWithMaxTokens(maxTokens int64) OpenAIOption {
	return func(c *OpenAI) {
		if maxTokens != 0 {
			c.maxTokens = maxTokens
		}
	}
}

// OpenAIWithSocksProxy sets the SOCKS proxy.
func OpenAIWithSocksProxy(socksProxy string) OpenAIOption {
	return func(c *OpenAI) {
		if socksProxy != "" {
			c.socksProxy = socksProxy
		}
	}
}

// OpenAIWithBaseURL sets the base API URL.
func OpenAIWithBaseURL(baseURL string) OpenAIOption {
	return func(c *OpenAI) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// NewOpenAI creates a new client instance with the given options.
func NewOpenAI(apiKey string, opts ...OpenAIOption) (*OpenAI, error) {
	c := &OpenAI{
		model:       openai.ChatModelGPT4o,
		temperature: 0.1,
		topP:        1.0,
		maxTokens:   1024,
	}

	for _, opt := range opts {
		opt(c)
	}

	openaiOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if c.baseURL != "" {
		openaiOpts = append(openaiOpts, option.WithBaseURL(c.baseURL))
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

		openaiOpts = append(openaiOpts, option.WithHTTPClient(&http.Client{Transport: transport}))
	}

	c.client = openai.NewClient(openaiOpts...)

	return c, nil
}

// CompleteChat implements the Completion interface.
func (c *OpenAI) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrOpenAIUnsupportedContentType,
				msg.Content,
			)
		}

		var message openai.ChatCompletionMessageParamUnion
		switch msg.Sender {
		case chat.RoleSystem:
			message = openai.SystemMessage(content)
		case chat.RoleUser:
			message = openai.UserMessage(content)
		case chat.RoleAssistant:
			message = openai.AssistantMessage(content)
		default:
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %v",
				ErrOpenAIUnsupportedRole,
				msg.Sender,
			)
		}

		openAIMessages[i] = message
	}

	chatCompletion, err := c.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:               openai.F(c.model),
			Messages:            openai.F(openAIMessages),
			Temperature:         openai.Float(c.temperature),
			TopP:                openai.Float(c.topP),
			MaxCompletionTokens: openai.Int(c.maxTokens),
		},
	)
	if err != nil {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: chatgpt request failed: %w",
			ErrOpenAIRequestFailed,
			err,
		)
	}

	if len(chatCompletion.Choices) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrOpenAIRequestFailed,
		)
	}

	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: chatCompletion.Choices[0].Message.Content,
	}, nil
}
