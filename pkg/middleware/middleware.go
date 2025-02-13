package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/dgrijalva/jwt-go"
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
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("secret_token")), nil
		})
		if err != nil {
			errttp.SendError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			username := claims["username"].(string)
			ctx := context.WithValue(r.Context(), UsernameKey, username)
			fmt.Println(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			errttp.SendError(w, http.StatusUnauthorized, "invalid token")
			return
		}
	}
}
