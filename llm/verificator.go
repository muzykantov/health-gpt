package llm

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
)

// ChatCompleter генерирует ответы с помощью языковой модели.
type ChatCompleter interface {
	CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error)
}

// Verificator проверяет ответы языковой модели с помощью дополнительных промптов и
// делает retry если есть подозрение на галлюцинирование модели.
type Verificator struct {
	orig ChatCompleter
}

// NewVerificator возвращает инициализированный верификатор.
func NewVerificator(orig ChatCompleter) *Verificator {
	return &Verificator{
		orig: orig,
	}
}

// CompleteChat запрашивает ответ от LLM и проверяет результат её работы.
func (v *Verificator) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	// TODO: Implement me.
	return v.orig.CompleteChat(ctx, msgs)
}
