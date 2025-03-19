package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/server"
)

type Command string

const (
	CmdUnspecified  Command = ""
	CmdClear        Command = "clear"
	CmdStart        Command = "start"
	CmdExit         Command = "exit"
	CmdMyGenetics   Command = "mygenetics"
	CmdMyGeneticsAI Command = "mygenetics_ai"
)

var commandsMessage = chat.MsgA(content.Commands{
	Items: []content.Command{
		{
			Name:        string(CmdStart),
			Description: "Начать новый диалог и предложить варианты",
		},
		{
			Name:        string(CmdClear),
			Description: "Начать новую сессию (забыть диалог)",
		},
		{
			Name:        string(CmdMyGenetics),
			Description: "Показать список анализов",
		},
		{
			Name:        string(CmdMyGeneticsAI),
			Description: "Показать список анализов с интерпретацией ИИ",
		},
		{
			Name:        string(CmdExit),
			Description: "Выйти из аккаунта",
		},
	},
})

func commands(cmd Command) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			w.WriteResponse(commandsMessage)

			switch cmd {
			case CmdUnspecified:
				return

			case CmdStart:
				greetings().Serve(ctx, w, r)

			case CmdClear:
				clear(true).Serve(ctx, w, r)

			case CmdExit:
				exit().Serve(ctx, w, r)

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
