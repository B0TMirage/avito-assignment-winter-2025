package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/jwtutils"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		errttp.SendError(w, http.StatusBadRequest, "invalid username or password")
		return
	}

	if user.Username == "" || user.Password == "" || len(user.Password) < 4 {
		errttp.SendError(w, http.StatusBadRequest, "invalid username or password")
		return
	}

	var storedPassword string
	err = database.DB.QueryRow("SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedPassword)

	if err == sql.ErrNoRows {
		// если пользователь не найден, регистрируем его
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			errttp.SendError(w, http.StatusInternalServerError, "couldn't process the password")
			return
		}
		_, err = database.DB.Exec("INSERT INTO users(username, password, coins) VALUES($1, $2, 1000)", user.Username, string(hashedPassword))
		if err != nil {
			errttp.SendError(w, http.StatusInternalServerError, "failed to register user")
			return
		}

	} else if err != nil {
		// прочие ошибки
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user info")
		return
	} else {
		// если найден, проверяем пароль
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password))
		if err != nil {
			errttp.SendError(w, http.StatusUnauthorized, "invalid password")
			return
		}
	}

	tokenString, err := jwtutils.CreateToken(user.Username, os.Getenv("secret_token"))
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func GetAllInfoHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	user.Username = r.Context().Value(middleware.UsernameKey).(string)
	err := database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	var allUserInfo models.UserInfoResponse
	allUserInfo.Coins = user.Coins

	// получение всех предметов пользователя с подсчётом их количества
	rows, err := database.DB.Query(`SELECT merch.title, COUNT(merch_id)
									FROM users_merch 
									JOIN merch ON merch_id = merch.id
									WHERE user_id = $1
									GROUP BY merch.title`, user.ID)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user's merch info")
		return
	}

	for rows.Next() {
		var userInventoryItem models.MerchInventoryItem

		rows.Scan(&userInventoryItem.Type, &userInventoryItem.Quantity)

		allUserInfo.Inventory = append(allUserInfo.Inventory, userInventoryItem)
	}

	var allTransactions models.Transaction

	// получение транзакций отправки coins
	rows, err = database.DB.Query(`SELECT users.username, SUM(amount) FROM transactions
								   JOIN users ON from_user_id = users.id WHERE to_user_id = $1
								   GROUP BY users.username`, user.ID)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user's transactions info")
		return
	}

	for rows.Next() {
		var receiveTransaction models.Receive

		rows.Scan(&receiveTransaction.FromUser, &receiveTransaction.Amount)

		allTransactions.Received = append(allTransactions.Received, receiveTransaction)
	}

	// получение транзакций получения coins
	rows, err = database.DB.Query(`SELECT users.username, SUM(amount) FROM transactions
								   JOIN users ON to_user_id = users.id WHERE from_user_id = $1
								   GROUP BY users.username`, user.ID)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user's transactions info")
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
	item := r.PathValue("item")

	if item == "" {
		errttp.SendError(w, http.StatusBadRequest, "incorrect item name")
		return
	}

	var merch models.Merch
	err := database.DB.QueryRow("SELECT id, title, price FROM merch WHERE title=$1", item).Scan(&merch.ID, &merch.Title, &merch.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			errttp.SendError(w, http.StatusBadRequest, "item not found")
			return
		}
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	var user models.User
	user.Username = r.Context().Value(middleware.UsernameKey).(string)
	err = database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	if user.Coins < merch.Price {
		errttp.SendError(w, http.StatusBadRequest, "insufficient funds")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "couldn't connect to the database")
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", merch.Price, user.ID)
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	_, err = tx.Exec("INSERT INTO users_merch(user_id, merch_id) VALUES($1, $2)", user.ID, merch.ID)
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendCoinsHandler(w http.ResponseWriter, r *http.Request) {
	sendCoinRequest := struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}{}
	json.NewDecoder(r.Body).Decode(&sendCoinRequest)
	if sendCoinRequest.ToUser == "" || sendCoinRequest.Amount <= 0 {
		errttp.SendError(w, http.StatusBadRequest, "invalid request")
		return
	}

	var toUserID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE username=$1", sendCoinRequest.ToUser).Scan(&toUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			errttp.SendError(w, http.StatusBadRequest, "user does not exist")
			return
		}
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	var user models.User
	user.Username = r.Context().Value(middleware.UsernameKey).(string)
	err = database.DB.QueryRow("SELECT id, coins FROM users WHERE username=$1", user.Username).Scan(&user.ID, &user.Coins)
	if err != nil {
		errttp.SendError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	if user.ID == toUserID {
		errttp.SendError(w, http.StatusBadRequest, "forbidden to send coins to yourself")
		return
	}

	if user.Coins < sendCoinRequest.Amount {
		errttp.SendError(w, http.StatusBadRequest, "insufficient funds")
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "couldn't connect to the database")
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2", sendCoinRequest.Amount, user.ID)
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	_, err = tx.Exec("UPDATE users SET coins = coins + $1 WHERE id = $2", sendCoinRequest.Amount, toUserID)
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	_, err = tx.Exec("INSERT INTO transactions(from_user_id, to_user_id, amount) VALUES($1, $2, $3)", user.ID, toUserID, sendCoinRequest.Amount)
	if err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		errttp.SendError(w, http.StatusInternalServerError, "error when querying the database")
		return
	}

	w.WriteHeader(http.StatusOK)
}
