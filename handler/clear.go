package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

func clear(response bool) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if err := r.Storage.SaveChatHistory(ctx, r.ChatID, []chat.Message{
				chat.MsgA("Начало диалога"),
			}); err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка записи истории чата: %v", err))
			}

			if response {
				w.WriteResponse(chat.MsgU("🧹 История чата очищена."))
			}
		},
	)
}
