package handler

import (
	"context"
	_ "embed"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/server"
)

//go:embed prompts/codelabs.txt
var MyGeneticsCodelabsPrompt string

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
				myGeneticsChat().Serve(ctx, w, r)

			case content.SelectItem:
				myGeneticsCodelab(msgContent.Data).Serve(ctx, w, r)

			case content.Command:
				commands(Command(msgContent.Name)).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.MsgA("⛔ Неизвестная команда. " +
					"Пожалуйста, выберите действие из предложенного списка."))
			}
		},
	)
}
