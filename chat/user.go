package chat

import "github.com/muzykantov/health-gpt/mygenetics"

// UserState определяет состояние пользователя.
type UserState int

// UserState определяет возможные состояния пользователя.
const (
	UserStateUnauthorized UserState = iota
	UserStateAuthorized
)

// User представляет пользователя.
type User struct {
	ID        int64
	FirstName string
	LastName  string
	UserName  string
	Email     string
	Password  string
	Tokens    []mygenetics.Token
	State     UserState
}
