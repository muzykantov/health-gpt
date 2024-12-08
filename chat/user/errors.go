package user

import "errors"

// Основные ошибки при работе с пользователем.
var (
	ErrUserNotFound = errors.New("user not found")
)
