package handler

import (
	"context"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

// myGeneticsCodelab создает обработчик для отображения результатов конкретного анализа.
// Если код начинается с "ai:", предоставляет интерпретацию через ИИ, в противном случае
// показывает детальные результаты. Требует авторизации пользователя.
func myGeneticsCodelab(code string) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			var useAI bool
			if strings.HasPrefix(code, "ai:") {
				code = strings.TrimPrefix(code, "ai:")
				useAI = true
			}

			w.WriteResponse(chat.MsgAf("🔍 Загружаю результаты анализа %s. "+
				"Это займёт несколько секунд...", code))

			features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, code)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось получить информацию об анализе. " +
					"Пожалуйста, попробуйте позже или обратитесь в поддержку."))

				r.ErrorLog.Printf("failed to fetch features (chatID: %d): %v", r.ChatID, err)
				return
			}

			if !useAI {
				for i, feature := range features {
					time.Sleep(time.Millisecond * 300)
					select {
					case <-ctx.Done():
						return

					default:
						w.WriteResponse(chat.MsgAf("%s\n📑 Показываю результат %d из %d.",
							feature, i+1, len(features)))
					}
				}

				return
			}

			msgs := make([]chat.Message, 0, len(features)+1)
			msgs = append(msgs, chat.MsgS(MyGeneticsCodelabsPrompt))

			for _, feature := range features {
				msgs = append(msgs, chat.MsgU(feature.String()))
			}

			w.WriteResponse(chat.MsgAf("📑 Загружено %d параметров анализа. "+
				"Приступаю к обработке...", len(features)))

			w.WriteResponse(chat.MsgA("⌛ Анализирую результаты с помощью ИИ. " +
				"Это может занять до минуты..."))

			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось получить интерпретацию результатов. " +
					"Пожалуйста, попробуйте позже или " +
					"просмотрите результаты без анализа ИИ."))

				r.ErrorLog.Printf("failed to complete chat (chatID: %d): %v", r.ChatID, err)
				return
			}

			w.WriteResponse(chat.MsgA(response.Content))
		},
	)
}
