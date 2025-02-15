package handlers

import (
	"database/sql"
	"net/http"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"
)

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
