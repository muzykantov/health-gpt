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

// MyGenetics создает основной обработчик для работы с генетическими анализами.
// Обрабатывает все типы запросов (команды, выбор анализа, текстовые сообщения)
// и маршрутизирует их к соответствующим обработчикам. Поддерживает историю чата.
func MyGenetics() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			history, err := r.Storage.GetChatHistory(ctx, r.ChatID, 1)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка получения истории чата: %v", err))
				return
			}

			if len(history) == 0 {
				myGeneticsGreetings().Serve(ctx, w, r)

				if err := r.Storage.SaveChatHistory(ctx, r.ChatID, []chat.Message{
					chat.MsgA("Начало диалога"),
				}); err != nil {
					w.WriteResponse(chat.MsgAf("⚠️ Ошибка записи истории чата: %v", err))
				}

				return
			}

			switch msgContent := r.Incoming.Content.(type) {
			case string:
				myGeneticsChat().Serve(ctx, w, r)

			case content.SelectItem:
				myGeneticsCodelab(msgContent.Data).Serve(ctx, w, r)

			case content.Command:
				myGeneticsCommands(CodelabsCommand(msgContent.Name)).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.MsgA("⛔ Неизвестная команда. " +
					"Пожалуйста, выберите действие из предложенного списка."))
			}
		},
	)
}
