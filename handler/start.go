package handler

import (
	"context"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/chat/content"
	"github.com/muzykantov/health-gpt/server"
)

func Start() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			if r.From.State == chat.UserStateUnauthorized {
				msg := chat.MsgA(content.Commands{
					Items: []content.Command{
						{
							Name:        string(CmdStart),
							Description: "Начать общение с ботом",
						},
					},
				})

				w.WriteResponse(msg)
			} else {
				commands(CmdUnspecified).Serve(ctx, w, r) // List of commands
			}

			auth(myGenetics()).Serve(ctx, w, r)
		},
	)
}
