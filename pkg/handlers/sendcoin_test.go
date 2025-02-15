package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
)

func TestSendCoin(t *testing.T) {
	url := "http://localhost:8080/api/sendCoin"

	token, err := authTestUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	database.DB.Exec("INSERT INTO users(username, password) VALUES('toTestUser', 'avito') ON CONFLICT (username) DO UPDATE SET coins = 0")

	var userID, toUserID int
	database.DB.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
	database.DB.QueryRow("SELECT id FROM users WHERE username = 'toTestUser'").Scan(&toUserID)
	database.DB.Exec("UPDATE users SET coins = 50 WHERE id = $1", userID)

	client := &http.Client{}

	reqData, err := json.Marshal(map[string]interface{}{"toUser": "toTestUser", "amount": 50})
	if err != nil {
		t.Error("error marshaling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		t.Fatal("error creating request: ", err)
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error making request: ", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	var userCoins, toUserCoins int
	database.DB.QueryRow("SELECT coins FROM users WHERE id = $1", userID).Scan(&userCoins)
	database.DB.QueryRow("SELECT coins FROM users WHERE id = $1", toUserID).Scan(&toUserCoins)

	if userCoins != 0 || toUserCoins != 50 {
		t.Fatal("money transaction error")
	}
}

func TestSendCoinNoMoney(t *testing.T) {
	url := "http://localhost:8080/api/sendCoin"

	token, err := authTestUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	database.DB.Exec("INSERT INTO users(username, password) VALUES('toTestUser', 'avito') ON CONFLICT (username) DO UPDATE SET coins = 0")
	var userID, toUserID int
	database.DB.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
	database.DB.QueryRow("SELECT id FROM users WHERE username = 'toTestUser'").Scan(&toUserID)
	database.DB.Exec("UPDATE users SET coins = 10 WHERE id = $1", userID)

	client := &http.Client{}

	reqData, err := json.Marshal(map[string]interface{}{"toUser": "toTestUser", "amount": 50})
	if err != nil {
		t.Error("error marshaling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		t.Fatal("error creating request: ", err)
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error making request: ", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}

	var userCoins, toUserCoins int
	database.DB.QueryRow("SELECT coins FROM users WHERE id = $1", userID).Scan(&userCoins)
	database.DB.QueryRow("SELECT coins FROM users WHERE id = $1", toUserID).Scan(&toUserCoins)

	if userCoins != 10 || toUserCoins != 0 {
		t.Fatal("money transaction error")
	}
}

func TestSendCoinIncorrectUser(t *testing.T) {
	url := "http://localhost:8080/api/sendCoin"

	token, err := authTestUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	var userID int
	database.DB.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
	database.DB.Exec("UPDATE users SET coins = 12000 WHERE id = $1", userID)

	client := &http.Client{}

	reqData, err := json.Marshal(map[string]interface{}{"toUser": "toMe", "amount": 10000})
	if err != nil {
		t.Error("error marshaling JSON:", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		t.Fatal("error creating request: ", err)
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error making request: ", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}

	var userCoins int
	database.DB.QueryRow("SELECT coins FROM users WHERE id = $1", userID).Scan(&userCoins)

	if userCoins != 12000 {
		t.Fatal("money transaction error")
	}
}
