package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

func SetupRoutes() {
	http.HandleFunc("POST /api/auth", LoginHandler)
	http.HandleFunc("POST /api/buy/{item}", middleware.AuthMiddleware(BuyHandler))
	http.HandleFunc("POST /api/sendCoin", middleware.AuthMiddleware(SendCointHandler))

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var storedPassword string
	err = database.DB.QueryRow("SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedPassword)

	if err == sql.ErrNoRows {
		// если пользователь не найден, регистрируем его
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
		}
		_, err = database.DB.Exec("INSERT INTO users(username, password, coins) VALUES($1, $2, 1000)", user.Username, string(hashedPassword))
		if err != nil {
			fmt.Println(err)
		}

	} else if err != nil {
		return
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"ttl":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("secret_token")))
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    tokenString,
		MaxAge:   int(time.Hour) * 24,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "username",
		Value:    user.Username,
		MaxAge:   int(time.Hour) * 24,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func GetAllInfoHandler(w http.ResponseWriter, r *http.Request) {

}

func BuyHandler(w http.ResponseWriter, r *http.Request) {
	item := strings.TrimPrefix(r.URL.Path, "/api/buy/")
	fmt.Println(item)

	if item == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usernameCookie, _ := r.Cookie("username")
	username := usernameCookie.Value
	var userID int
	var userCoins int
	database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", username).Scan(&userID, &userCoins)

	var merch models.Merch
	database.DB.QueryRow("SELECT id, title, price FROM merch WHERE title=$1", item).Scan(&merch.ID, &merch.Title, &merch.Price)

	if userCoins < merch.Price {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", merch.Price, userID)
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT INTO users_merch(user_id, merch_id) VALUES($1, $2)", userID, merch.ID)
	if err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendCointHandler(w http.ResponseWriter, r *http.Request) {

}
