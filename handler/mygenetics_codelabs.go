package handler

import (
	"context"
	"fmt"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

func myGeneticsCodelabs(cmd Command) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			codelabs, err := mygenetics.DefaultClient.FetchCodelabs(ctx, access)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось загрузить список анализов. " +
					"Пожалуйста, попробуйте позже или обратитесь в поддержку."))
				return
			}

			if len(codelabs) == 0 {
				w.WriteResponse(chat.MsgA("⚠️ У вас пока нет доступных анализов. " +
					"Новые результаты появятся здесь автоматически."))
				return
			}

			var (
				cmdMyGenetics   = cmd == CmdMyGenetics || cmd == CmdUnspecified
				cmdMyGeneticsAI = cmd == CmdMyGeneticsAI || cmd == CmdUnspecified
			)

			if cmdMyGenetics {
				msgContent := content.Select{
					Header: "🧪 Выберите анализ, чтобы просмотреть детальные результаты:",
				}
				for _, codelab := range codelabs {
					msgContent.Items = append(msgContent.Items, content.SelectItem{
						Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
						Data:    PrefixCodelab + codelab.Code,
					})
				}

				w.WriteResponse(chat.MsgA(msgContent))
			}

			if cmdMyGeneticsAI {
				msgContent := content.Select{
					Header: "🧪 Выберите анализ для получения " +
						"развёрнутой интерпретации результатов с помощью ИИ:",
				}
				for _, codelab := range codelabs {
					msgContent.Items = append(msgContent.Items, content.SelectItem{
						Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
						Data:    PrefixAI + codelab.Code,
					})
				}

				w.WriteResponse(chat.MsgA(msgContent))
			}
		},
	)
}
