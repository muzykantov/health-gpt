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
			cookieStr:   "PHPSESSID=z3cknROZuReLQH3r3FsL0eMdFht8tDar; path=/; httponly; samesite=lax",
			wantType:    "PHPSESSID",
			wantExpires: "0001-01-01 00:00:00 UTC",
		},
		{
			name:        "Valid Access Token",
			cookieStr:   "accessToken=MjUwMWM4Njc4ZTAwZjk3N2FhMGRkZGI3NDFhMzgxZWI3M2U2Y2JmYTUwY2UzNGYzNmMxMTExYjk0ZjkxNTBkZWMwZTMyZGVhZGUwNGYyODdlMTQwNTc1MDI4YmY2YzI4ZDM2Mzk5ZDU1Zjc4YzFlODQ3NWZmYjFjNDc2NDlhNGVkZWE1ZjZmMTk1OTBjNTc5ZDE5YzJkYWVjNGNiMGRiNmNmZDNhMjlkNDExYjU0ZDRhZDMxNjYyMmMyZWRmZjBhM2M2N2QyODI3YWU4MTBlZWYxMjk3Yjg4ZjdkNmYyN2Y2NWZiZDQ2M2Y1YWUyYWYyMDgwMDIzM2VjZjI4Y2NlN2EwMDY4M2YzODBlMGYzNDY4ZmM2Zjk4YWQxODBkNTc5OGRjYjI4MTk4NTgzYmZlYjk2NWY0YmU4ODM5N2NmZGI0YTgyMWE3MDgzM2QyZTAyZDhhYjA2MzgzZmQ0MWE1ZDVkM2FiZGZiZDA5MzI5MTE5NDIwYmQ0Y2I3Y2VkMzE1NGI1Njc5ZDllY2FiYjMwZjIyYTRmNGJmYmI1YmY5ZDBjNzNlNzk3NTdkMDk3Mzk2MjM0NWY0ZTU4NTkyNTY4Y2YyOWMyZjc3YzIyMzhmMTRlMjhmMDY4YjAwZDFmNmFmOWE5YzA3YWU4OGVjODM2MzM2NjhjNjY4NjcyMmYzYTg0ZWM4Y2MxM2YwYWZlNjY3YjE1NTlhNzMyNDA2YjE4ZTg4ZTA4MDllMDJjZWFiZDUyODU4YjhkY2UwYTA2YTYwZTY4MjgzZDAwZGU4NTI4MDU3ZWQ2NTc5OTc3ZGJjNmE2MzI5ZmM3OTAxMTBjM2M5M2I5MjMyMGVhOTAxMzc4NmI3MTBmYmYzYTQzMDkwZTQ1NWNjNmNlZGUyZmJiYTNhMWJkZmQ4OTQxZjY0YzZhYWU2Y2Y3NDAyOGZhNGZhZGI5M2RhNzZjMjM1Y2E4Nw%253D%253D; expires=Wed, 05 Mar 2025 20:59:03 GMT; Max-Age=3600; path=/; domain=.mygenetics.ru; secure; httponly; samesite=lax",
			wantType:    "accessToken",
			wantExpires: "2025-03-05 20:59:03 GMT",
		},
		{
			name:        "Valid Refresh Token",
			cookieStr:   "refreshToken=YWI5ZDI1YWIwNmM2NmNiYmU5NDg3NDJmOTUyYzNiMTgwYzBlNGZjMWM3ZDdjZjllMmVmMjkyZTFkOTlmOGMyZWRjZmZmMjlkYmEyNDE0ZWJkNGZiZDdlNWM5NWY0YjA1NTZjMDU5MDViNzJjM2EyY2Q0ZWE2Mjc5NDhkMTUyMGJlMGFlZWQyNDZlYWE3N2MwMjJjM2I0ZDc5NjhhMmM3MWUwNDJlNzM3ODJhYWQwOTA4MDc3MmEwMjE3NTk2MzlhZjBjODhkMzAyZTUxMzhmNjQ5MTlmZTU3MzEzN2ZlYjE2ZmZlMzVjZjJiMmI2YjQxZDUxMWU1NzM5NzhmODkxMWI4MDRjN2MwMGRmYmIzY2RlN2E2ZDkzZmNiOTFmNjVlZWEzNWU3YjE4OWQxOTRkNzk3N2M3MGEyMWUxYWViOGUzOGY5Njg3ZGY0ODg3YjBlNjVlOGE5Mzk5NTBmYTcwNzNlM2I3NjI1ZDQxY2MxOWI2ZTljMDU0ZGIyZjVlMTRjY2NjMmNmYmI2MDRjYmExOGM1YWQyZjNlOTNhZDBmNzYzZDY1MTYzYTZkOGVkYjg5; expires=Thu, 06 Mar 2025 23:59:03 GMT; Max-Age=100800; path=/; domain=.mygenetics.ru; secure; httponly; samesite=lax",
			wantType:    "refreshToken",
			wantExpires: "2025-03-06 23:59:03 GMT",
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
