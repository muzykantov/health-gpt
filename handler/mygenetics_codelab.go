package handler

import (
	"context"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/handler/prompts"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const myGeneticsCodelabPrompt = "codelab"

// myGeneticsCodelab создает обработчик для отображения результатов конкретного анализа.
// Если код начинается с "ai:", предоставляет интерпретацию через ИИ, в противном случае
// показывает детальные результаты. Требует авторизации пользователя.
func myGeneticsCodelab(data SelectItemData) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			var useAI bool
			switch {
			case strings.HasPrefix(data, PrefixAI):
				data = strings.TrimPrefix(data, PrefixAI)
				useAI = true
			case strings.HasPrefix(data, PrefixCodelab):
				data = strings.TrimPrefix(data, PrefixCodelab)
				useAI = false
			default:
				w.WriteResponse(chat.MsgAf("⛔ Неизвестный префикс: %s.", data))
				return
			}

			w.WriteResponse(chat.MsgAf("🔍 Загружаю результаты анализа %s. "+
				"Это займёт несколько секунд...", data))

			features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, data)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось получить информацию об анализе. " +
					"Пожалуйста, попробуйте позже или обратитесь в поддержку."))

				r.Log.Printf("failed to fetch features (chatID: %d): %v", r.ChatID, err)
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
							feature.ToHTML(), i+1, len(features)))
					}
				}

				return
			}

			prompt := prompts.Get(myGeneticsCodelabPrompt, r.Completer.ModelName())
			if prompt == prompts.Default {
				w.WriteResponse(chat.MsgA("⛔ Промпт не найден."))
				return
			}

			msgs := make([]chat.Message, 0, 2)
			msgs = append(msgs, chat.MsgS(prompt))
			msgs = append(msgs, chat.MsgU(features.BuildLLMContext()))

			w.WriteResponse(chat.MsgAf("📑 Загружено %d параметров анализа. "+
				"Приступаю к обработке...", len(features)))

			w.WriteResponse(chat.MsgA("⌛ Анализирую результаты с помощью ИИ. " +
				"Это может занять до минуты..."))

			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось получить интерпретацию результатов. " +
					"Пожалуйста, попробуйте позже или " +
					"просмотрите результаты без анализа ИИ."))

				r.Log.Printf("failed to complete chat (chatID: %d): %v", r.ChatID, err)
				return
			}

			w.WriteResponse(chat.MsgA(response.Content))
		},
	)
}
