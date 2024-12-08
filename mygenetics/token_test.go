package mygenetics

import "testing"

func TestParseToken(t *testing.T) {
	testCases := []struct {
		name        string
		cookieStr   string
		wantType    string
		wantExpires string
		wantErr     bool
	}{
		{
			name:        "Valid PHP Token",
			cookieStr:   "PHPSESSID=Z50gVUDr66A34TILV5YCCssijanHzjLx; path=/; HttpOnly",
			wantType:    "PHPSESSID",
			wantExpires: "0001-01-01 00:00:00 UTC",
		},
		{
			name:        "Valid Access Token",
			cookieStr:   "accessToken=YWRlYjE3NGM3ZDhiZDIxMDBlNTIxNDJiY2Y3YWY1NzFmYzg0ZTA1OWM3OWVhOTE2YzlkZTlmYmU4NmMyZmRlOGFmZGJkZWFhMWM2M2EyNDAxN2ZiMDYwMmQ5OWY0M2MzMzc0NTIzOTE0YzBiZjVlNTEzZjllZWI4M2RlM2ZmN2YxOWEwNmU3NzVkOWRkZTVlYjU4NWE1ZGQ2MjkyMjMyNTg2ZWVlNTBjZTMxZTUyYmNjODFjOTk1ZWRiMzNkYWY2OWJhOGE1MWJmMmQwNjlkZTA4ZmNiMDU5MzYwNmJmMWUxYjEzNjJkMmE3NDY4ZWVmYmMwODA1NzFmN2JkODk2OWUzNWM2M2Q1NzBhZjkxYWNjNGI4MTVjNGNiNWMxZjU1NjJhNWQ1OWFjMjM5NTg2NGQ1NGNhYjBmZmY3MDViMTI1YmQ3YjEwMDhjZmQ4OWM2MDNiZTcwNmI5YTA0ZTkwNGMzYzQxNWUwMGE1Mzg5NjBiMGY4ZmRlYzI5NTJmMmIyNzhiOTdlYTNlZWNkOGM3NzA4NzQzZTkzNmRkY2RiMmEyN2QwNDgzYmFkMjczNmJjOWI1ODQ1MGQ5MTEyMmJmZWYyOTBlMWU2YmYwNzNiMDRlMGE3ZTc0NzI3ODFjYWY5Y2E2OTk0NGI4ZjZjZTNiMDcyNDk1NzRjMTVhMTY5ZDBmZjIzNWY1YjUxNGQ1MTI4YjM2OGIzN2JjMmM2MmU3NTk5ZDVkNDY1MDk2ZDc3N2YyOGZjZWQ3NDk4NjdiY2UzYjI5Y2FmN2Q4MzUwNTM2NTk3MjU0MTQyZmY2MmQ4YjJlMmExN2RjYjU4YmRiMmVjOTY3ZmNmMDNhNGFiZGM5MGQyMzZkZmU3MTUxYmQwMjQyZDY4NjA4ODk2Y2JkODgyNjliYzk5NmU4YWM4NDIzY2U4NmIyNzdlZWM1OTgzMWVhNDdmMTFhMTYzZmMzMzJkMzQyYTMwNGMxYWJhNDE%3D; expires=Mon, 28-Oct-2024 14:55:36 GMT; Max-Age=3600; path=/; domain=.mygenetics.ru; secure; HttpOnly",
			wantType:    "accessToken",
			wantExpires: "2024-10-28 14:55:36 GMT",
		},
		{
			name:        "Valid Refresh Token",
			cookieStr:   "refreshToken=OTRlMTIyYmYzMGIwOGRhNGJiMzYzYWNkYjU0NDU5OTcyZTRhOWJjZWMzYzhkY2VhNmEyZDhkMGZhYWE0OWFhNmQwNWFmMWE4ZWU2ZDc4MjE0MDMxMTAyMGI0NzVjZTBmMDQxOWQ0ZGZmMzMyNzhmMzc2MmMxMjBkN2U4NGQ2NWM4NzY2ZjdmMTc5Y2IzZWM1YjIxNzYwYTAyMGM2YjkyNzY1ODI3MDhjNzYzMGQ3MmJiYThlZDIzYzIxM2NlMWEyNTgwNDg3N2QyMmExYWU0ODRiNGMzZmQxYTkwNzg3MTY3ZDk0NmU1NTczMTMxYWVkNGFkYmMxN2RjYmQ1YTJjZjIwOTlkMjgwM2M2ZWNlMmI2YWZjNTUzNDI2MThhOGY1MWM0YmI0N2IwZGZkNTQwZTBiNThkOGEwMmU2MmE4Mjg1ZDE3MTE0NzEwNDVhZmQwNGEyYWM5ZGE3MjJhN2Q2YjA2ZTliZjI2MGFjMGVlYzQ0ZWQyNTg5ZTQ5YmQ4MDE5NmQ4OTJmNjlhZmEwNTc2ZDI1ZjBjY2FjOGYyNzZjNjRkNzc0NWNlYjI1OWRiZWM5; expires=Tue, 29-Oct-2024 17:55:36 GMT; Max-Age=100800; path=/; domain=.mygenetics.ru; secure; HttpOnly",
			wantType:    "refreshToken",
			wantExpires: "2024-10-29 17:55:36 GMT",
		},
		{
			name:        "Invalid Cookie Format",
			cookieStr:   "invalidcookie",
			wantType:    "",
			wantExpires: "0001-01-01 00:00:00 UTC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := Token(tc.cookieStr)

			if typ := token.Type(); typ != tc.wantType {
				t.Errorf("Token type = %v, want %v", typ, tc.wantType)
			}

			gotExpires := token.Expires().Format("2006-01-02 15:04:05 MST")
			if gotExpires != tc.wantExpires {
				t.Errorf(
					"Token expires = %v, want %v",
					gotExpires,
					tc.wantExpires,
				)
			}
		})
	}
}
