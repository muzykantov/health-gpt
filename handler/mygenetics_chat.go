package handler

import (
	"context"
	_ "embed"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

//go:embed prompts/chat.txt
var myGeneticsChatPrompt string

// myGeneticsChat создает обработчик для чата с ИИ по вопросам генетических анализов.
// Обрабатывает текстовые сообщения пользователя и предоставляет ответы на основе
// всех доступных результатов анализов. Требует авторизации пользователя.
func myGeneticsChat() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			msgText, ok := r.Incoming.Content.(string)
			if !ok {
				w.WriteResponse(chat.MsgA("⛔ Пожалуйста, отправьте текстовое сообщение."))
				r.ErrorLog.Printf("invalid message content type (chatID: %d): expected string, got %T",
					r.ChatID, r.Incoming.Content)
				return
			}

			access := mygenetics.AccessToken(r.From.Tokens)
			if access == "" {
				w.WriteResponse(chat.MsgA("⚠️ Для доступа к анализам необходимо авторизоваться. " +
					"Пожалуйста, введите свой email и пароль."))
				return
			}

			history, err := r.History.ReadChatHistory(ctx, r.ChatID, 100)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка получения истории чата: %v", err))
				r.ErrorLog.Printf("failed to read chat history (chatID: %d): %v", r.ChatID, err)
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
				r.ErrorLog.Printf("failed to fetch codelabs (chatID: %d): %v", r.ChatID, err)
				return
			}

			if len(codelabs) == 0 {
				w.WriteResponse(chat.MsgA("⚠️ У вас пока нет доступных анализов. " +
					"Пожалуйста, загрузите анализы, чтобы начать общение."))
				return
			}

			var allFeatures []string
			for _, codelab := range codelabs {
				features, err := mygenetics.DefaultClient.FetchFeatures(ctx, access, codelab.Code)
				if err != nil {
					w.WriteResponse(chat.MsgAf("⚠️ Не удалось загрузить результаты анализа %s: %v",
						codelab.Code, err))
					r.ErrorLog.Printf("failed to fetch features for codelab %s (chatID: %d): %v",
						codelab.Code, r.ChatID, err)
					continue
				}
				for _, feature := range features {
					allFeatures = append(allFeatures, feature.String())
				}
			}

			// Формируем сообщения для AI:
			// системный промпт в начале + все анализы + история чата + новое сообщение
			msgs := make([]chat.Message, 0, 1+len(allFeatures)+len(filteredHistory)+1)
			msgs = append(msgs, chat.MsgS(myGeneticsChatPrompt))
			for _, feature := range allFeatures {
				msgs = append(msgs, chat.MsgU(feature))
			}
			msgs = append(msgs, filteredHistory...)
			msgs = append(msgs, chat.MsgU(msgText))

			// w.WriteResponse(chat.MsgA("🤔 Анализирую ваш вопрос..."))

			done := make(chan struct{})
			go func() {
				ticker := time.NewTicker(time.Second)
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
				r.ErrorLog.Printf("failed to complete chat (chatID: %d): %v", r.ChatID, err)
				return
			}

			done <- struct{}{}

			// Сохраняем всю историю плюс новые сообщения
			newHistory := make([]chat.Message, len(history)+2)
			copy(newHistory, history)
			newHistory[len(history)] = chat.MsgU(msgText)
			newHistory[len(history)+1] = chat.MsgA(response.Content)

			if err := r.History.WriteChatHistory(ctx, r.ChatID, newHistory); err != nil {
				w.WriteResponse(chat.MsgAf("⚠️ Ошибка сохранения истории чата: %v", err))
				r.ErrorLog.Printf("failed to write chat history (chatID: %d): %v", r.ChatID, err)
				return
			}

			w.WriteResponse(chat.MsgA(response.Content))
		},
	)
}
