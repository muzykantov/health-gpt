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

// DeepSeekOption определяет функцию-опцию для конфигурации клиента.
type DeepSeekOption func(*DeepSeek)

// DeepSeek реализует клиент для взаимодействия с DeepSeek API.
type DeepSeek struct {
	client *http.Client
	// Конфигурационные параметры.
	model       string  // Модель для использования.
	temperature float64 // Температура генерации (0.0-2.0).
	topP        float64 // Top-p сэмплирование (0.0-1.0).
	maxTokens   int64   // Максимальное количество токенов в ответе.
	socksProxy  string  // Адрес прокси-сокса.
	baseURL     string  // Базовый URL для API.
	apiKey      string  // API ключ для авторизации.
}

// Установка модели для использования.
func DeepSeekWithModel(model string) DeepSeekOption {
	return func(c *DeepSeek) {
		c.model = model
	}
}

// Установка температуры генерации.
func DeepSeekWithTemperature(temperature float64) DeepSeekOption {
	return func(c *DeepSeek) {
		c.temperature = temperature
	}
}

// Установка параметра top-p сэмплирования.
func DeepSeekWithTopP(topP float64) DeepSeekOption {
	return func(c *DeepSeek) {
		c.topP = topP
	}
}

// Установка максимального количества токенов.
func DeepSeekWithMaxTokens(maxTokens int64) DeepSeekOption {
	return func(c *DeepSeek) {
		c.maxTokens = maxTokens
	}
}

// Установка SOCKS прокси.
func DeepSeekWithSocksProxy(socksProxy string) DeepSeekOption {
	return func(c *DeepSeek) {
		c.socksProxy = socksProxy
	}
}

// Установка базового URL для API.
func DeepSeekWithBaseURL(baseURL string) DeepSeekOption {
	return func(c *DeepSeek) {
		c.baseURL = baseURL
	}
}

// NewDeepSeek создает новый экземпляр DeepSeek с заданными опциями.
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

// DeepSeekMessage представляет сообщение в формате DeepSeek API.
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest представляет запрос к DeepSeek API.
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Stream      bool              `json:"stream"`
	Temperature float64           `json:"temperature"`
	TopP        float64           `json:"top_p"`
	MaxTokens   int64             `json:"max_tokens"`
}

// DeepSeekResponse представляет ответ от DeepSeek API.
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

// CompleteChat реализует интерфейс Completion.
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

	// Формирование запроса
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
