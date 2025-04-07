package handler

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/handler/prompts"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const authPrompt = "auth"

func auth(next server.Handler) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			// Если токен не истек, то пользователь авторизован, передаем запрос дальше.
			if mygenetics.AccessToken(r.From.Tokens).Expires().After(time.Now().Add(time.Minute * 5)) {
				next.Serve(ctx, w, r)
				return
			}

			// Если есть email и пароль, то авторизуем пользователя (email и пароль проверены).
			if r.From.Email != "" && r.From.Password != "" {
				var err error

				if r.From.Tokens, err = mygenetics.DefaultClient.Authenticate(
					ctx,
					r.From.Email,
					r.From.Password,
				); err != nil {
					w.WriteResponse(chat.MsgAf("⛔ Ошибка аутентификации mygenetics: %v", err))
					return
				}

				if err := r.Storage.SaveUser(ctx, r.From); err != nil {
					w.WriteResponse(chat.MsgAf("⛔ Ошибка обновления информации о пользователе: %v", err))
					return
				}

				next.Serve(ctx, w, r)
				return
			}

			// Если пользователь ешё не ввел email и пароль, то читаем историю чата.
			msgs, err := r.Storage.GetChatHistory(ctx, r.ChatID, 0)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка получения истории чата: %v", err))
				return
			}

			// Пользователь ноывй? Добавляем инстукции для ИИ получить email и пароль.
			if len(msgs) == 0 {
				prompt := prompts.Get(authPrompt, r.Completer.ModelName())
				if prompt == prompts.Default {
					w.WriteResponse(chat.MsgA("⛔ Промпт не найден."))
					return
				}

				msgs = []chat.Message{
					chat.MsgS(prompt),
				}
			}

			if _, ok := r.Incoming.Content.(string); !ok {
				r.Incoming.Content = "Привет."
			}

			// Добавляем присланное пользователем сообщение в контекст.
			msgs = append(msgs, r.Incoming)

			// Даем ИИ разобраться с сообщениями и решить что делать дальше.
			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка генерации ответа: %v", err))
				return
			}

			// Если ИИ не нашел в переписке email и пароль, отправляем ответ ИИ пользователю.
			if !json.Valid([]byte(response.Content.(string))) {
				msgs = append(msgs, response)

				if err := r.Storage.SaveChatHistory(ctx, r.ChatID, msgs); err != nil {
					w.WriteResponse(chat.MsgAf("⛔ Ошибка сохранения истории чата: %v", err))
					return
				}

				w.WriteResponse(response)
				return
			}

			// Если ИИ ответил JSON-ом, то значит он нашел email и пароль.
			var credentials struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			if err := json.Unmarshal(
				[]byte(response.Content.(string)),
				&credentials,
			); err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка парсинга ответа: %v", err))
				return
			}

			// Пробуем авторизовать пользователя.
			tokens, err := mygenetics.DefaultClient.Authenticate(
				ctx,
				credentials.Email,
				credentials.Password,
			)
			if err != nil {
				// Если не получилось, то сбрасываем переписку и отправляем ответ пользователю.
				w.WriteResponse(
					chat.MsgA("❌ Имя пользователя или пароль не подходят. Попробуйте ещё раз."),
				)

				if err := r.Storage.SaveChatHistory(
					ctx,
					r.ChatID,
					make([]chat.Message, 0),
				); err != nil {
					w.WriteResponse(chat.MsgAf("⛔ Ошибка сохранения истории чата: %v", err))
				}
				return
			}

			// Если получилось, то сохраняем данные пользователя.
			r.From.Email = credentials.Email
			r.From.Password = credentials.Password
			r.From.Tokens = tokens
			r.From.State = chat.UserStateAuthorized
			if err := r.Storage.SaveUser(ctx, r.From); err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка сохранения пользователя: %v", err))
				return
			}

			// Сбрасываем переписку.
			if err := r.Storage.SaveChatHistory(
				ctx,
				r.ChatID,
				make([]chat.Message, 0),
			); err != nil {
				w.WriteResponse(chat.MsgAf("⛔ Ошибка сохранения истории чата: %v", err))
				return
			}

			w.WriteResponse(chat.MsgA("✅ Вы успешно вошли в систему! Благодарим за предоставленные данные."))

			// Передаем запрос дальше.
			r.Incoming = chat.NewMessage(chat.RoleUser, "")
			next.Serve(ctx, w, r)
		},
	)
}
