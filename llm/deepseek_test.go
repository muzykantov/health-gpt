package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/muzykantov/health-gpt/chat"

	"github.com/stretchr/testify/require"
)

func TestDeepSeekCompletion_Integration(t *testing.T) {
	// Пропуск теста, если DEEPSEEK_API_KEY не установлен.
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}

	opts := []DeepSeekOption{
		DeepSeekWithTemperature(0.7),
		DeepSeekWithMaxTokens(256),
	}

	if proxy := os.Getenv("DEEPSEEK_SOCKS_PROXY"); proxy != "" {
		opts = append(opts, DeepSeekWithSocksProxy(proxy))
	}

	// Инициализация DeepSeek.
	client, err := NewDeepSeek(apiKey, opts...)
	require.NoError(t, err, "Failed to create DeepSeek client")

	// Создание контекста с таймаутом.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Подготовка запроса.
	messages := []chat.Message{
		{
			Sender:  chat.RoleSystem,
			Content: "Ты - полезный ассистент. Отвечай кратко, в одно предложение.",
		},
		{
			Sender:  chat.RoleUser,
			Content: "Какая столица России?",
		},
	}

	// Отправка запроса.
	response, err := client.CompleteChat(ctx, messages)
	require.NoError(t, err, "Failed to complete request")

	// Проверка ответа.
	require.Equal(t, chat.RoleAssistant, response.Sender)

	// Проверка текста ответа.
	content, ok := response.Content.(string)
	require.True(t, ok, "Response content should be string")
	require.NotEmpty(t, content, "Response text should not be empty")

	t.Logf("Response: %s", content)
}
