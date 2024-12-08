package mygenetics

import (
	"strings"
	"time"
)

type Token string

func (t Token) Expires() (exp time.Time) {
	const (
		field  = "expires="
		format = "Mon, 02 Jan 2006 15:04:05 MST" // Fri, 06 Dec 2024 20:30:39 GMT
	)

	if pos := strings.Index(string(t), field); pos != -1 {
		exp, _ = time.Parse(
			format,
			string(t[pos+len(field):pos+len(field)+len(format)]),
		)
	}

	return
}

func (t Token) Type() (typ string) {
	if pos := strings.Index(string(t), "="); pos != -1 {
		typ = string(t[:pos])
	}

	return
}

func AccessToken(tokens []Token) Token {
	for _, token := range tokens {
		if token.Type() == "accessToken" {
			return token
		}
	}

	return ""
}

func RefreshToken(tokens []Token) Token {
	for _, token := range tokens {
		if token.Type() == "refreshToken" {
			return token
		}
	}

	return ""
}
