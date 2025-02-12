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
	http.HandleFunc("GET /api/info", middleware.AuthMiddleware(GetAllInfoHandler))
	http.HandleFunc("POST /api/auth", LoginHandler)
	http.HandleFunc("GET /api/buy/{item}", middleware.AuthMiddleware(BuyHandler))
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
	usernameCookie, _ := r.Cookie("username")
	var user models.User
	user.Username = usernameCookie.Value
	err := database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)
	if err != nil {
		return
	}

	var allUserInfo models.UserInfoResponse
	allUserInfo.Coins = user.Coins

	rows, err := database.DB.Query(`SELECT merch.title, COUNT(merch_id)
									FROM users_merch 
									JOIN merch ON merch_id = merch.id
									WHERE user_id = $1
									GROUP BY merch.title`, user.ID)
	if err != nil {
		return
	}

	for rows.Next() {
		var userInventoryItem models.MerchInventoryItem

		rows.Scan(&userInventoryItem.Type, &userInventoryItem.Quantity)

		allUserInfo.Inventory = append(allUserInfo.Inventory, userInventoryItem)
	}

	var allTransactions models.Transaction
	rows, err = database.DB.Query(`SELECT users.username, SUM(amount) FROM transactions
								   JOIN users ON from_user_id = users.id WHERE to_user_id = $1
								   GROUP BY users.username`, user.ID)
	if err != nil {
		return
	}

	for rows.Next() {
		var receiveTransaction models.Receive

		rows.Scan(&receiveTransaction.FromUser, &receiveTransaction.Amount)

		allTransactions.Received = append(allTransactions.Received, receiveTransaction)
	}

	rows, err = database.DB.Query(`SELECT users.username, SUM(amount) FROM transactions
								   JOIN users ON to_user_id = users.id WHERE from_user_id = $1
								   GROUP BY users.username`, user.ID)
	if err != nil {
		return
	}

	for rows.Next() {
		var sentTransaction models.Sent

		rows.Scan(&sentTransaction.ToUser, &sentTransaction.Amount)

		allTransactions.Sent = append(allTransactions.Sent, sentTransaction)
	}

	allUserInfo.CoinHistory = allTransactions

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(allUserInfo)
}

func BuyHandler(w http.ResponseWriter, r *http.Request) {
	item := strings.TrimPrefix(r.URL.Path, "/api/buy/")

	if item == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usernameCookie, _ := r.Cookie("username")

	var user models.User
	user.Username = usernameCookie.Value
	database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)

	var merch models.Merch
	database.DB.QueryRow("SELECT id, title, price FROM merch WHERE title=$1", item).Scan(&merch.ID, &merch.Title, &merch.Price)

	if user.Coins < merch.Price {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", merch.Price, user.ID)
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT INTO users_merch(user_id, merch_id) VALUES($1, $2)", user.ID, merch.ID)
	if err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendCointHandler(w http.ResponseWriter, r *http.Request) {
	sendCoinRequest := struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}{}
	json.NewDecoder(r.Body).Decode(&sendCoinRequest)
	if sendCoinRequest.ToUser == "" || sendCoinRequest.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var toUserID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE username=$1", sendCoinRequest.ToUser).Scan(&toUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		return
	}

	usernameCookie, _ := r.Cookie("username")

	var user models.User
	user.Username = usernameCookie.Value
	err = database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)
	if err != nil {
		return
	}

	if user.Coins < sendCoinRequest.Amount {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", sendCoinRequest.Amount, user.ID)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins + $1 WHERE id = $2", sendCoinRequest.Amount, toUserID)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO transactions(from_user_id, to_user_id, amount) VALUES($1, $2, $3)", user.ID, toUserID, sendCoinRequest.Amount)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
