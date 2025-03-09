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
	ErrChatGPTUnsupportedRole        = errors.New("unsupported role")
	ErrChatGPTUnsupportedContentType = errors.New("unsupported content type")
	ErrChatGPTRequestFailed          = errors.New("request failed")
)

// ChatGPTOption defines a configuration function for the client.
type ChatGPTOption func(*ChatGPT)

// ChatGPT implements a client for interacting with ChatGPT.
type ChatGPT struct {
	client *openai.Client
	// Configuration parameters.
	model       openai.ChatModel // Model to use.
	temperature float64          // Generation temperature (0.0-2.0).
	topP        float64          // Top-p sampling (0.0-1.0).
	maxTokens   int64            // Maximum number of tokens in response.
	socksProxy  string           // SOCKS proxy address.
	baseURL     string           // Base API URL.
}

// ChatGPTWithModel sets the model to use.
func ChatGPTWithModel(model string) ChatGPTOption {
	return func(c *ChatGPT) {
		if model != "" {
			c.model = model
		}
	}
}

// ChatGPTWithTemperature sets the generation temperature.
func ChatGPTWithTemperature(temperature float64) ChatGPTOption {
	return func(c *ChatGPT) {
		if temperature != 0 {
			c.temperature = temperature
		}
	}
}

// ChatGPTWithTopP sets the top-p sampling parameter.
func ChatGPTWithTopP(topP float64) ChatGPTOption {
	return func(c *ChatGPT) {
		if topP != 0 {
			c.topP = topP
		}
	}
}

// ChatGPTWithMaxTokens sets the maximum number of tokens.
func ChatGPTWithMaxTokens(maxTokens int64) ChatGPTOption {
	return func(c *ChatGPT) {
		if maxTokens != 0 {
			c.maxTokens = maxTokens
		}
	}
}

// ChatGPTWithSocksProxy sets the SOCKS proxy.
func ChatGPTWithSocksProxy(socksProxy string) ChatGPTOption {
	return func(c *ChatGPT) {
		if socksProxy != "" {
			c.socksProxy = socksProxy
		}
	}
}

// ChatGPTWithBaseURL sets the base API URL.
func ChatGPTWithBaseURL(baseURL string) ChatGPTOption {
	return func(c *ChatGPT) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// NewChatGPT creates a new ChatGPT instance with the given options.
func NewChatGPT(apiKey string, opts ...ChatGPTOption) (*ChatGPT, error) {
	c := &ChatGPT{
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
func (c *ChatGPT) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
	for i, msg := range msgs {
		content, ok := msg.Content.(string)
		if !ok {
			return chat.EmptyMessage, fmt.Errorf(
				"%w: %T",
				ErrChatGPTUnsupportedContentType,
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
				ErrChatGPTUnsupportedRole,
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
			ErrChatGPTRequestFailed,
			err,
		)
	}

	if len(chatCompletion.Choices) == 0 {
		return chat.EmptyMessage, fmt.Errorf(
			"%w: no choices in response",
			ErrChatGPTRequestFailed,
		)
	}

	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: chatCompletion.Choices[0].Message.Content,
	}, nil
}
