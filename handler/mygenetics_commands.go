package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

// myGeneticsCommands создает обработчик для выполнения команд работы с генетическими анализами.
// При получении CmdMyGenetics или CmdMyGeneticsAI делегирует обработку в myGeneticsCodelabs,
// иначе возвращает сообщение об ошибке.
func myGeneticsCommands(cmd CodelabsCommand) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			switch cmd {
			case CmdMyGenetics:
				myGeneticsCodelabs(CmdMyGenetics).Serve(ctx, w, r)

			case CmdMyGeneticsAI:
				myGeneticsCodelabs(CmdMyGeneticsAI).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.MsgA("⛔ Неизвестная команда. " +
					"Пожалуйста, выберите действие из предложенного списка."))
			}
		},
	)
}
