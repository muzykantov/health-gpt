package handler

import (
	"context"
	_ "embed"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/handler/prompts"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const myGeneticsChatPrompt = "chat"

// myGeneticsChat создает обработчик для чата с ИИ по вопросам генетических анализов.
// Обрабатывает текстовые сообщения пользователя и предоставляет ответы на основе
// всех доступных результатов анализов. Требует авторизации пользователя.
func myGeneticsChat() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			msgText, ok := r.Incoming.Content.(string)
			if !ok {
				w.WriteResponse(chat.MsgA("⛔ Пожалуйста, отправьте текстовое сообщение."))
				r.Log.Printf("invalid message content type (chatID: %d): expected string, got %T",
					r.ChatID, r.Incoming.Content)
				return
			}

			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			history, err := r.Storage.GetChatHistory(ctx, r.ChatID, 100)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка получения истории чата: %v", err))
				r.Log.Printf("failed to read chat history (chatID: %d): %v", r.ChatID, err)
				return
			}

			// Фильтруем историю только для отправки в AI
			var filteredHistory []chat.Message
			for _, msg := range history {
				if text, ok := msg.Content.(string); ok {
					msg.Content = text
					filteredHistory = append(filteredHistory, msg)
				}
			}

			// w.WriteResponse(chat.MsgA("🔍 Загружаю ваши анализы..."))

			codelabs, err := mygenetics.DefaultClient.FetchCodelabs(ctx, access)
			if err != nil {
				w.WriteResponse(chat.MsgA("⚠️ Не удалось загрузить анализы. " +
					"Пожалуйста, попробуйте позже или обратитесь в поддержку."))
				r.Log.Printf("failed to fetch codelabs (chatID: %d): %v", r.ChatID, err)
				return
			}

			if len(codelabs) == 0 {
				w.WriteResponse(chat.MsgA("⚠️ У вас пока нет доступных анализов. " +
					"Пожалуйста, загрузите анализы, чтобы начать общение."))
				return
			}

			featureSet, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, codelabs[0].Code)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Не удалось загрузить результаты анализа %s: %v",
					codelabs[0].Code, err))
				r.Log.Printf("failed to fetch features for codelab %s (chatID: %d): %v",
					codelabs[0].Code, r.ChatID, err)
			}
			// var featureSet genetics.FeatureSet
			// for _, codelab := range codelabs {
			// 	features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, codelab.Code)
			// 	if err != nil {
			// 		w.WriteResponse(chat.MsgAf("⚠️ Не удалось загрузить результаты анализа %s: %v",
			// 			codelab.Code, err))
			// 		r.Log.Printf("failed to fetch features for codelab %s (chatID: %d): %v",
			// 			codelab.Code, r.ChatID, err)
			// 		continue
			// 	}

			// 	featureSet = featureSet.MergeWith(features)
			// }

			// -----------------------------------------------------------------
			// Формирование контекста AI:
			// -----------------------------------------------------------------

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
			confirmationMsg := "Я изучил предоставленные генетические данные. Теперь я готов ответить на ваши вопросы, опираясь на эту информацию."
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

			w.WriteResponse(chat.MsgA(response.Content))
		},
	)
}
