package handler

import (
	"context"
	"fmt"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

// Chat - простая LLM для тестирования.
func Chat() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if _, ok := r.Incoming.Content.(string); !ok {
				return
			}

			history, err := r.History.ReadChatHistory(ctx, r.ChatID, 10)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprintf("History error: %v", err),
					),
				)
				return
			}

			if len(history) == 0 {
				history = append(
					history,
					chat.NewMessage(
						chat.RoleSystem,
						"Ты - полезный ассистент. Отвечай кратко, в одно предложение.",
					),
				)
			}

			history = append(history, r.Incoming)

			response, err := r.Completer.CompleteChat(ctx, history)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprintf("AI error: %v", err),
					),
				)
				return
			}

			history = append(history, response)

			if err := r.History.WriteChatHistory(ctx, r.ChatID, history); err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprintf("History error: %v", err),
					),
				)
				return
			}

			w.WriteResponse(response)
		},
	)
}
