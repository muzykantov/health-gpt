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

// ChatGPTOption определяет функцию-опцию для конфигурации клиента.
type ChatGPTOption func(*ChatGPT)

// ChatGPT реализует клиент для взаимодействия с ChatGPT.
type ChatGPT struct {
	client *openai.Client
	// Конфигурационные параметры.
	model       openai.ChatModel // Модель для использования.
	temperature float64          // Температура генерации (0.0-2.0).
	topP        float64          // Top-p сэмплирование (0.0-1.0).
	maxTokens   int64            // Максимальное количество токенов в ответе.
	socksProxy  string           // Адрес прокси-сокса.
	baseURL     string           // Базовый URL для API.
}

// Установка модели для использования.
func ChatGPTWithModel(model string) ChatGPTOption {
	return func(c *ChatGPT) {
		c.model = model
	}
}

// Установка температуры генерации.
func ChatGPTWithTemperature(temperature float64) ChatGPTOption {
	return func(c *ChatGPT) {
		c.temperature = temperature
	}
}

// Установка параметра top-p сэмплирования.
func ChatGPTWithTopP(topP float64) ChatGPTOption {
	return func(c *ChatGPT) {
		c.topP = topP
	}
}

// Установка максимального количества токенов.
func ChatGPTWithMaxTokens(maxTokens int64) ChatGPTOption {
	return func(c *ChatGPT) {
		c.maxTokens = maxTokens
	}
}

func ChatGPTWithSocksProxy(socksProxy string) ChatGPTOption {
	return func(c *ChatGPT) {
		c.socksProxy = socksProxy
	}
}

// Установка базового URL для API.
func ChatGPTWithBaseURL(baseURL string) ChatGPTOption {
	return func(c *ChatGPT) {
		c.baseURL = baseURL
	}
}

// NewChatGPT создает новый экземпляр ChatGPT с заданными опциями.
func NewChatGPT(apiKey string, opts ...ChatGPTOption) (*ChatGPT, error) {
	// Инициализация клиента со значениями по умолчанию.
	c := &ChatGPT{
		model:       openai.ChatModelGPT4o,
		temperature: 0.1,
		topP:        1.0,
		maxTokens:   1024,
	}

	// Применение переданных опций.
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

// CompleteChat реализует интерфейс Completion.
func (c *ChatGPT) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	// Преобразование сообщений в формат OpenAI.
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

	// Запрос к ChatGPT API.
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

	// Преобразование ответа в нужный формат.
	return chat.Message{
		Sender:  chat.RoleAssistant,
		Content: chatCompletion.Choices[0].Message.Content,
	}, nil
}
