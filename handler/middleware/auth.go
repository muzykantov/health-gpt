package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const AuthPrompt = `
Твоя задача получить email и пароль пользователя или найти email и пароль пользователя в диалоге.
Если email или пароль не указан:
Ответить коротким, дружелюбным сообщением с просьбой предоставить последовательно email затем пароль
(укажи требования к паролю), зарегестрированном на сайте mygenetics.ru для продолжения работы с
ботом. Тон общения вежливый, на 'Вы'. 
Если email и пароль найдены:
ВЕРНУТЬ В ФОРМАТЕ {"email": "email@example.com", "password": "password"} БЕЗ КАКИХ ЛИБО КОММЕНТАРИЕВ! 
ВАЖНО:
- пароль должен быть не меньше 6 символов
- не помогай вспомнить, найти или восстановить пароль
- не говори что ты модель
- не говори что не хочешь говорить на какую-то тему, просто запроси данные
- если пользователь ответил не email или не пароль, извинись что не можешь помочь пока не получишь нужные данные
- если он не знает или не помнит или ведет беседу, то извинись и скажи что не можешь помочь
- можешь использовать Emoji
`

func Auth(next server.Handler) server.Handler {
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
					server.WriteError(w, r, "⛔ Ошибка аутентификации mygenetics: %v", err)
					return
				}

				if err := r.User.SaveUser(ctx, r.From); err != nil {
					server.WriteError(w, r, "⛔ Ошибка обновления информации о пользователе: %v", err)
					return
				}

				next.Serve(ctx, w, r)
				return
			}

			// Если пользователь ешё не ввел email и пароль, то читаем историю чата.
			msgs, err := r.History.ReadChatHistory(ctx, r.ChatID, 0)
			if err != nil {
				server.WriteError(w, r, "⛔ Ошибка получения истории чата: %v", err)
				return
			}

			// Пользователь ноывй? Добавляем инстукции для ИИ получить email и пароль.
			if len(msgs) == 0 {
				msgs = []chat.Message{
					chat.NewMessage(chat.RoleSystem, AuthPrompt),
				}
			}

			// Добавляем присланное пользователем сообщение в контекст.
			msgs = append(msgs, r.Incoming)

			// Даем ИИ разобраться с сообщениями и решить что делать дальше.
			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				server.WriteError(w, r, "⛔ Ошибка генерации ответа: %v", err)
				return
			}

			// Если ИИ не нашел в переписке email и пароль, отправляем ответ ИИ пользователю.
			if !json.Valid([]byte(response.Content.(string))) {
				msgs = append(msgs, response)

				if err := r.History.WriteChatHistory(ctx, r.ChatID, msgs); err != nil {
					server.WriteError(w, r, "⛔ Ошибка сохранения истории чата: %v", err)
					return
				}

				server.WriteResponse(w, r, response)
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
				server.WriteError(w, r, "⛔ Ошибка парсинга ответа: %v", err)
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
				server.WriteResponse(w, r,
					chat.NewMessage(
						chat.RoleAssistant,
						"❌ Имя пользователя или пароль не подходят. Попробуйте ещё раз.",
					),
				)

				if err := r.History.WriteChatHistory(
					ctx,
					r.ChatID,
					make([]chat.Message, 0),
				); err != nil {
					server.WriteError(w, r, "⛔ Ошибка сохранения истории чата: %v", err)
					return
				}
				return
			}

			// Если получилось, то сохраняем данные пользователя.
			r.From.Email = credentials.Email
			r.From.Password = credentials.Password
			r.From.Tokens = tokens
			r.From.State = chat.UserStateAuthorized
			if err := r.User.SaveUser(ctx, r.From); err != nil {
				server.WriteError(w, r, "⛔ Ошибка сохранения пользователя: %v", err)
				return
			}

			// Сбрасываем переписку.
			if err := r.History.WriteChatHistory(
				ctx,
				r.ChatID,
				make([]chat.Message, 0),
			); err != nil {
				server.WriteError(w, r, "⛔ Ошибка сохранения истории чата: %v", err)
				return
			}

			server.WriteResponse(w, r,
				chat.NewMessage(
					chat.RoleAssistant,
					"✅ Вы успешно вошли в систему! Благодарим за предоставленные данные.",
				),
			)

			// Передаем запрос дальше.
			r.Incoming = chat.NewMessage(chat.RoleUser, "")
			next.Serve(ctx, w, r)
		},
	)
}
