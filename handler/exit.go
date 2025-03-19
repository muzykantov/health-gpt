package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

func exit() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if err := r.Storage.SaveChatHistory(ctx, r.ChatID, []chat.Message{}); err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка записи истории чата: %v", err))
			}

			r.From.Password = ""
			r.From.Tokens = nil
			r.From.State = chat.UserStateUnauthorized

			if err := r.Storage.SaveUser(ctx, r.From); err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка сохранения пользователя: %v", err))
				return
			}

			w.WriteResponse(chat.MsgA("👋 Вы успешно вышли из системы. До новых встреч!"))

		},
	)
}
