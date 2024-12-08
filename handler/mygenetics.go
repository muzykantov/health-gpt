package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

func MyGenetics() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			switch content := r.Incoming.Content.(type) {
			case string:
				myGeneticsCodelabs().Serve(ctx, w, r)

			case chat.SelectContentItem:
				myGeneticsCodelab(content.Data).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.NewMessage(chat.RoleAssistant, "⛔ Неизвестная команда."))
			}
		},
	)
}

func myGeneticsCodelabs() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			access := mygenetics.AccessToken(r.From.Tokens)

			codelabs, err := mygenetics.DefaultClient.FetchCodelabs(ctx, access)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprint("⛔ Ошибка получения списка анализов: ", err),
					),
				)
			}

			if len(codelabs) == 0 {
				w.WriteResponse(
					chat.NewMessage(chat.RoleAssistant, "⚠️ Список анализов пуст. Попробуйте позже."),
				)

				return
			}

			content := chat.SelectContent{
				Header: "🧪 Выберите анализ для отправки результатов в чат:",
			}
			for _, codelab := range codelabs {
				content.Items = append(content.Items, chat.SelectContentItem{
					Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
					Data:    codelab.Code,
				})
			}

			w.WriteResponse(
				chat.NewMessage(chat.RoleAssistant, content),
			)

			content = chat.SelectContent{
				Header: "🧪  Выберите анализ для заключения ИИ:",
			}
			for _, codelab := range codelabs {
				content.Items = append(content.Items, chat.SelectContentItem{
					Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
					Data:    "ai:" + codelab.Code,
				})
			}

			w.WriteResponse(
				chat.NewMessage(chat.RoleAssistant, content),
			)
		},
	)
}

func myGeneticsCodelab(code string) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			defer myGeneticsCodelabs().Serve(ctx, w, r)

			access := mygenetics.AccessToken(r.From.Tokens)

			if strings.HasPrefix(code, "ai:") {
				code = strings.TrimPrefix(code, "ai:")
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprintf("🧪 Запрашиваю ИИ интерпретацию результатов анализа %s...", code),
					),
				)

				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						"⚠️ ИИ не отвечает...",
					),
				)

				return
			}

			w.WriteResponse(
				chat.NewMessage(
					chat.RoleAssistant,
					fmt.Sprintf("🧪 Запрашиваю результаты анализа %s...", code),
				),
			)

			features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, code)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						fmt.Sprint("⛔ Ошибка получения информации об анализе: ", err),
					),
				)
			}

			for i, feature := range features {
				time.Sleep(time.Second)

				select {
				case <-ctx.Done():
					return

				default:
					w.WriteResponse(
						chat.NewMessage(
							chat.RoleAssistant,
							feature.String()+
								"\n"+
								fmt.Sprintf(
									"📑 Признак %d из %d.", i+1, len(features),
								),
						),
					)
				}
			}
		},
	)
}
