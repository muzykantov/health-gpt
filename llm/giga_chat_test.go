package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/muzykantov/health-gpt/chat"

	"github.com/stretchr/testify/require"
)

func TestGigaChatCompletion_Integration(t *testing.T) {
	// Пропуск теста, если GIGACHAT_CLIENT_ID или GIGACHAT_CLIENT_SECRET не установлены.
	clientID := os.Getenv("GIGACHAT_CLIENT_ID")
	clientSecret := os.Getenv("GIGACHAT_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("GIGACHAT_CLIENT_ID or GIGACHAT_CLIENT_SECRET not set")
	}

	// Инициализация GigaChat.
	client, err := NewGigaChat(
		clientID,
		clientSecret,
		GigaChatWithTemperature(0.7),
		GigaChatWithMaxTokens(256),
		GigaChatWithRepetitionPenalty(1.1),
	)
	require.NoError(t, err, "Failed to create GigaChat client")

	// Создание контекста с таймаутом.
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
