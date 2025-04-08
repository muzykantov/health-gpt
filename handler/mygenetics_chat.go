package handler

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/handler/prompts"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const (
	myGeneticsChatPrompt = "chat"
	codelabCodeCacheKey  = "chat_codelab_code:%d"
)

// myGeneticsChat создает обработчик для чата с ИИ по вопросам генетических анализов.
// Обрабатывает текстовые сообщения пользователя и предоставляет ответы на основе
// всех доступных результатов анализов. Требует авторизации пользователя.
func myGeneticsChat(data SelectItemData) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			var (
				codelabCodeCacheKey = fmt.Sprintf(codelabCodeCacheKey, r.ChatID)
				codelabCode         string
				msgText             string
				sendCode            bool
			)

			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			switch {
			case data == "":
				var ok bool

				msgText, ok = r.Incoming.Content.(string)
				if !ok {
					w.WriteResponse(chat.MsgA("⛔ Пожалуйста, отправьте текстовое сообщение."))
					r.Log.Printf("invalid message content type (chatID: %d): expected string, got %T",
						r.ChatID, r.Incoming.Content)
					return
				}

				if cachedCodelab, ok := r.Cache.Get(codelabCodeCacheKey); ok {
					codelabCode = cachedCodelab.(string)
					sendCode = true
					break
				}

				codelabs, err := mygenetics.DefaultClient.FetchCodelabs(ctx, access)
				if err != nil {
					w.WriteResponse(chat.MsgA("⚠️ Не удалось загрузить анализы. " +
						"Пожалуйста, попробуйте позже или обратитесь в поддержку."))
					r.Log.Printf("failed to fetch codelabs (chatID: %d): %v", r.ChatID, err)
					return
				}

				switch len(codelabs) {
				case 0:
					w.WriteResponse(chat.MsgA("⚠️ У вас пока нет доступных анализов. " +
						"Пожалуйста, загрузите анализы, чтобы начать общение."))
					return

				case 1:
					codelabCode = codelabs[0].Code
					sendCode = false

				default:
					msgContent := content.Select{
						Header: "🧬 У вас несколько анализов. Пожалуйста, выберите один. " +
							"Ассистент будет использовать его, пока выбор не сбросится " +
							"(например, автоматически или командой /clear).",
					}
					for _, codelab := range codelabs {
						msgContent.Items = append(msgContent.Items, content.SelectItem{
							Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
							Data: fmt.Sprintf(
								"%s%s:%s",
								PrefixAIChat,
								codelab.Code,
								r.Incoming.ID,
							),
						})
					}

					r.Cache.Add(PrefixAIChat+r.Incoming.ID, msgText)

					w.WriteResponse(chat.MsgA(msgContent))
					return
				}

			case strings.HasPrefix(data, PrefixAIChat):
				parts := strings.SplitN(strings.TrimPrefix(data, PrefixAIChat), ":", 2)
				if len(parts) != 2 {
					w.WriteResponse(chat.MsgA("⛔ Неверный формат анализа. Пожалуйста, выберите ответ из списка."))
					r.Log.Printf("invalid message parts (chatID: %d): %v",
						r.ChatID, parts)
					return
				}

				msg, ok := r.Cache.Get(PrefixAIChat + parts[1])
				if !ok {
					w.WriteResponse(chat.MsgA("⛔ Сообщение устарело."))
					r.Log.Printf("invalid message cache id (chatID: %d): %s",
						r.ChatID, parts[1])
					return
				}

				codelabCode = parts[0]
				msgText = msg.(string)
				sendCode = true

				r.Cache.Add(codelabCodeCacheKey, codelabCode)
				if err := r.Storage.SaveChatHistory(ctx, r.ChatID, []chat.Message{
					chat.MsgU(DefaultFirstMessage),
				}); err != nil {
					w.WriteResponse(chat.MsgAf("⚠️ Ошибка сохранения истории чата: %v", err))
					r.Log.Printf("failed to write chat history (chatID: %d): %v", r.ChatID, err)
					return
				}

			default:
				w.WriteResponse(chat.MsgA("⛔ Неизвестная команда. " +
					"Пожалуйста, выберите действие из предложенного списка."))
				return
			}

			featureSet, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, codelabCode)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Не удалось загрузить результаты анализа %s: %v",
					codelabCode, err))
				r.Log.Printf("failed to fetch features for codelab %s (chatID: %d): %v",
					codelabCode, r.ChatID, err)
				return
			}

			/*
				var featureSet genetics.FeatureSet
				for _, codelab := range codelabs {
					features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, codelab.Code)
					if err != nil {
						w.WriteResponse(chat.MsgAf("⚠️ Не удалось загрузить результаты анализа %s: %v",
							codelab.Code, err))
						r.Log.Printf("failed to fetch features for codelab %s (chatID: %d): %v",
							codelab.Code, r.ChatID, err)
						continue
					}

					featureSet = featureSet.MergeWith(features)
				}
			*/

			// -----------------------------------------------------------------
			// Формирование контекста AI:
			// -----------------------------------------------------------------

			history, err := r.Storage.GetChatHistory(ctx, r.ChatID, 100)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка получения истории чата: %v", err))
				r.Log.Printf("failed to read chat history (chatID: %d): %v", r.ChatID, err)
				return
			}

			var filteredHistory []chat.Message
			for _, msg := range history {
				if text, ok := msg.Content.(string); ok {
					if text == DefaultFirstMessage {
						continue
					}

					msg.Content = text
					filteredHistory = append(filteredHistory, msg)
				}
			}

			prompt := prompts.Get(myGeneticsChatPrompt, r.Completer.ModelName())
			if prompt == prompts.Default {
				w.WriteResponse(chat.MsgA("⛔ Промпт не найден."))
				return
			}

			contextMsg := "Следующие данные генетического анализа должны использоваться для ответа на мои вопросы:\n\n" +
				featureSet.BuildLLMContext() +
				"\n\nТеперь я буду задавать вопросы, опираясь на эти данные."

			msgs := make([]chat.Message, 0, 3+len(filteredHistory))
			msgs = append(msgs, chat.MsgS(prompt))     // Системный промпт
			msgs = append(msgs, chat.MsgU(contextMsg)) // Данные как сообщение пользователя

			// Подтверждающий ответ ассистента после контекста
			confirmationMsg := "Я изучил предоставленные генетические данные. " +
				"Теперь я готов ответить на ваши вопросы, опираясь на эту информацию."
			msgs = append(msgs, chat.MsgA(confirmationMsg))

			// История чата
			msgs = append(msgs, filteredHistory...)

			// И текущий вопрос пользователя
			msgs = append(msgs, chat.MsgU(msgText))

			// -----------------------------------------------------------------

			w.WriteResponse(chat.MsgA("🤔 Анализирую ваш вопрос..."))

			done := make(chan struct{})
			go func() {
				w.WriteResponse(chat.MsgA(content.Typing{}))
				ticker := time.NewTicker(time.Second * 10)
				for {
					select {
					case <-ticker.C:
						w.WriteResponse(chat.MsgA(content.Typing{}))
					case <-done:
						return
					}
				}
			}()

			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось получить ответ. " +
					"Пожалуйста, попробуйте позже или переформулируйте вопрос."))
				r.Log.Printf("failed to complete chat (chatID: %d): %v", r.ChatID, err)
				return
			}

			done <- struct{}{}

			// Сохраняем всю историю плюс новые сообщения
			newHistory := make([]chat.Message, len(history)+2)
			copy(newHistory, history)
			newHistory[len(history)] = chat.MsgU(msgText)
			newHistory[len(history)+1] = chat.MsgA(response.Content)

			if err := r.Storage.SaveChatHistory(ctx, r.ChatID, newHistory); err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка сохранения истории чата: %v", err))
				r.Log.Printf("failed to write chat history (chatID: %d): %v", r.ChatID, err)
				return
			}

			if !sendCode {
				w.WriteResponse(chat.MsgA(response.Content))
			} else {
				w.WriteResponse(chat.MsgAf(
					"🧠 Вот, что показывают данные из анализа %s.\n\n%s",
					codelabCode,
					response.Content,
				))
			}
		},
	)
}
