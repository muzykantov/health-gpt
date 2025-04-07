package llm

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
)

type Mock struct {
	CompleteChatFn func(ctx context.Context, msgs []chat.Message) (chat.Message, error)
}

func (m *Mock) ModelName() string {
	return "mock"
}

func (m *Mock) CompleteChat(ctx context.Context, msgs []chat.Message) (chat.Message, error) {
	return m.CompleteChatFn(ctx, msgs)
}
