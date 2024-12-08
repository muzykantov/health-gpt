package server

import (
	"fmt"

	"github.com/muzykantov/health-gpt/chat"
)

// WriteResponse пишет ответ на запрос.
func WriteResponse(w ResponseWriter, r *Request, m chat.Message) {
	if err := w.WriteResponse(m); err != nil {
		r.ErrorLog.Printf("error writing response failed: %v", err)
	}
}

// WriteError пишет ошибку в ответ на запрос.
func WriteError(w ResponseWriter, r *Request, format string, args ...any) {
	if err := w.WriteResponse(
		chat.NewMessage(
			chat.RoleAssistant,
			fmt.Sprintf(format, args...),
		),
	); err != nil {
		r.ErrorLog.Printf("writing response error failed: %v", err)
	}
}
