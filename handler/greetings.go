package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/server"
)

// greetings создает обработчик для отображения приветственного сообщения
// с доступными командами и инструкциями по работе с генетическими анализами.
// Отображается при первом входе пользователя в чат.
func greetings() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			w.WriteResponse(chat.MsgA("👋 Добро пожаловать! Вы можете выбрать анализы из " +
				"списка и получить их интерпретацию с помощью искусственного интеллекта. " +
				"Также вы можете задавать вопросы относительно имеющихся анализов в базе."))

			clear(false).Serve(ctx, w, r)
			commands(CmdUnspecified).Serve(ctx, w, r)
			myGeneticsCodelabs(CmdUnspecified).Serve(ctx, w, r)
		},
	)
}
