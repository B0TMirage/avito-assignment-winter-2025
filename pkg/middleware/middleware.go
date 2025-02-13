package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/jwtutils"
)

type contextKey string

const UsernameKey contextKey = "username"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			errttp.SendError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		claims, err := jwtutils.ValidateJWT(tokenString, os.Getenv("secret_token"))
		if err != nil {
			errttp.SendError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		username := claims["username"].(string)
		ctx := context.WithValue(r.Context(), UsernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
