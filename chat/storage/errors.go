package storage

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUnsupportedContentType = errors.New("unsupported content type")
)
