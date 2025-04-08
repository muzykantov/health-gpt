package handler

import (
	"context"
	"fmt"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

const (
	DefaultFirstMessage = "[Начало диалога]"
)

func clear(response bool) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if err := r.Storage.SaveChatHistory(ctx, r.ChatID, []chat.Message{
				chat.MsgU(DefaultFirstMessage),
			}); err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка записи истории чата: %v", err))
			}

			r.Cache.Remove(fmt.Sprintf(codelabCodeCacheKey, r.ChatID))

			if response {
				w.WriteResponse(chat.MsgU("🧹 История чата очищена."))
			}
		},
	)
}
