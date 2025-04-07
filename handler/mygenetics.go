package handler

import (
	"context"
	_ "embed"
	"strings"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/server"
)

type (
	SelectItemPrefix = string
	SelectItemData   = string
)

const (
	PrefixCodelab SelectItemPrefix = "codelab:"
	PrefixAI      SelectItemPrefix = "ai:"
	PrefixAIChat  SelectItemPrefix = "ai_chat:"
)

// myGenetics создает основной обработчик для работы с генетическими анализами.
// Обрабатывает все типы запросов (команды, выбор анализа, текстовые сообщения)
// и маршрутизирует их к соответствующим обработчикам. Поддерживает историю чата.
func myGenetics() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if r.From.State == chat.UserStateUnauthorized {
				w.WriteResponse(chat.MsgA("⛔ Пользователь не авторизован."))
				return
			}

			history, err := r.Storage.GetChatHistory(ctx, r.ChatID, 1)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка получения истории чата: %v", err))
				return
			}

			if len(history) == 0 {
				greetings().Serve(ctx, w, r)
				return
			}

			switch msgContent := r.Incoming.Content.(type) {
			case string:
				myGeneticsChat("").Serve(ctx, w, r)

			case content.SelectItem:
				switch {
				case strings.HasPrefix(msgContent.Data, PrefixAI):
					myGeneticsCodelab(msgContent.Data).Serve(ctx, w, r)

				case strings.HasPrefix(msgContent.Data, PrefixCodelab):
					myGeneticsCodelab(msgContent.Data).Serve(ctx, w, r)

				case strings.HasPrefix(msgContent.Data, PrefixAIChat):
					myGeneticsChat(msgContent.Data).Serve(ctx, w, r)
				}

			case content.Command:
				commands(Command(msgContent.Name)).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.MsgA("⛔ Неизвестная команда. " +
					"Пожалуйста, выберите действие из предложенного списка."))
			}
		},
	)
}
