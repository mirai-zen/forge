package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// JwtAuth 创建 JWT 鉴权中间件
// userEndpoint: user-service 的 /api/user/verify 地址
// 白名单路径会被跳过鉴权
func JwtAuth(userEndpoint string) func(http.HandlerFunc) http.HandlerFunc {
	skipPaths := []string{
		"/health",
		"/api/user/register",
		"/api/user/login",
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 白名单跳过
			for _, p := range skipPaths {
				if strings.HasPrefix(r.URL.Path, p) {
					next(w, r)
					return
				}
			}

			token := extractBearer(r)
			if token == "" {
				http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
				return
			}

			// 调 user-service 校验 token
			resp, err := http.Post(userEndpoint+"/api/user/verify",
				"application/json",
				strings.NewReader(`{"token":"`+token+`"}`))
			if err != nil || resp.StatusCode != http.StatusOK {
				http.Error(w, `{"error":"auth failed"}`, http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			// 透传验证结果到下游（注入 header）
			if bytes.Contains(body, []byte(`"valid":true`)) {
				next(w, r)
				return
			}

			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		}
	}
}

func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) < 7 || auth[:7] != "Bearer " {
		return ""
	}
	return auth[7:]
}
