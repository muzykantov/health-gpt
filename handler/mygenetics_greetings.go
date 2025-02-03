package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/server"
)

// myGeneticsGreetings создает обработчик для отображения приветственного сообщения
// с доступными командами и инструкциями по работе с генетическими анализами.
// Отображается при первом входе пользователя в чат.
func myGeneticsGreetings() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			commands := content.Commands{
				Items: []content.Command{
					{
						Name:        string(CmdMyGenetics),
						Description: "Показать список анализов",
					},
					{
						Name:        string(CmdMyGeneticsAI),
						Description: "Показать список анализов с интерпретацией ИИ",
					},
				},
			}

			w.WriteResponse(chat.MsgA(commands))

			w.WriteResponse(chat.MsgA("👋 Добро пожаловать! Вы можете выбрать анализы из " +
				"списка и получить их интерпретацию с помощью искусственного интеллекта. " +
				"Также вы можете задавать вопросы относительно имеющихся анализов в базе."))

			myGeneticsCodelabs(CmdUnspecified).Serve(ctx, w, r)
		},
	)
}
