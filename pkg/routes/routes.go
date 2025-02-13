package routes

import (
	"net/http"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/handlers"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"

	_ "github.com/lib/pq"
)

func SetupRouter() *http.ServeMux {
	mux := http.ServeMux{}

	mux.HandleFunc("GET /api/info", middleware.AuthMiddleware(handlers.GetAllInfoHandler))
	mux.HandleFunc("POST /api/auth", handlers.LoginHandler)
	mux.HandleFunc("GET /api/buy/{item}", middleware.AuthMiddleware(handlers.BuyHandler))
	mux.HandleFunc("POST /api/sendCoin", middleware.AuthMiddleware(handlers.SendCoinsHandler))

	return &mux
}
