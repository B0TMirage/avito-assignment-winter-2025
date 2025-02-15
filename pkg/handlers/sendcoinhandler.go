package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"

	_ "github.com/lib/pq"
)

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
