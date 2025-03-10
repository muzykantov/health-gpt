package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/muzykantov/health-gpt/chat"

	"github.com/stretchr/testify/require"
)

func TestOpenAICompletion_Integration(t *testing.T) {
	// Пропуск теста, если OPENAI_API_KEY не установлен.
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	opts := []OpenAIOption{
		OpenAIWithTemperature(0.7),
		OpenAIWithMaxTokens(256),
	}

	if proxy := os.Getenv("OPENAI_SOCKS_PROXY"); proxy != "" {
		opts = append(opts, OpenAIWithSocksProxy(proxy))
	}

	// Инициализация OpenAI.
	client, err := NewOpenAI(apiKey, opts...)
	require.NoError(t, err, "Failed to create OpenAI client")

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
	require.True(t, ok, "Response content should be TextContent")
	require.NotEmpty(t, content, "Response text should not be empty")

	t.Logf("Response: %s", content)
}
