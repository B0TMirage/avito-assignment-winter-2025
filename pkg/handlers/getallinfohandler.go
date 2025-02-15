package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/middleware"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"
)

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
		var item models.MerchInventoryItem

		rows.Scan(&item.Type, &item.Quantity)

		allUserInfo.Inventory = append(allUserInfo.Inventory, item)
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
