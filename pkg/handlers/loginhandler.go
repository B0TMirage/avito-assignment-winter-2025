package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/errttp"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/jwtutils"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"
	"golang.org/x/crypto/bcrypt"
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
