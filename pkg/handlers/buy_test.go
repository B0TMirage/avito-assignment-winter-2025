package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/models"

	_ "github.com/lib/pq"
)

var localDBURL string = "postgres://avito:passvito@127.0.0.1:5432/avito-merch-db?sslmode=disable"

func TestBuyHandler(t *testing.T) {
	token, err := authUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	// получаем id, задаём пользователю 250 монет и удаляем все предметы, которые могли быть до этого
	var id int
	database.DB.QueryRow(`SELECT id FROM users WHERE username=$1`, "testuser").Scan(&id)
	database.DB.Exec(`UPDATE users SET coins = 250 WHERE id=$1`, id)
	database.DB.Exec("DELETE FROM users_merch WHERE user_id=$1", id)

	client := &http.Client{}

	items := []string{"pen", "umbrella", "socks", "cup", "socks"}

	for _, v := range items {
		url := fmt.Sprint("http://localhost:8080/api/buy/", v)
		req, err := http.NewRequest("GET", url, nil)
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
	}

	rows, err := database.DB.Query("SELECT merch.title, COUNT(merch_id) FROM users_merch JOIN merch ON merch_id = merch.id WHERE user_id = $1 GROUP BY merch.title", id)
	if err != nil {
		t.Fatal("error querying db: ", err)
	}

	userInventory := []models.MerchInventoryItem{}
	for rows.Next() {
		var item models.MerchInventoryItem
		rows.Scan(&item.Type, &item.Quantity)
		userInventory = append(userInventory, item)
	}

	if len(userInventory) != 4 {
		t.Fatalf("got %v elements, want 4", len(userInventory))
	}

	want := map[string]int{"pen": 1, "umbrella": 1, "socks": 2, "cup": 1}

	for _, v := range userInventory {
		if want[v.Type] != v.Quantity {
			t.Fatal("error in the number of items")
		}
	}
}

func TestBuyHandlerNoMoney(t *testing.T) {
	token, err := authUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	// получаем id, задаём пользователю 0 монет
	var id int
	database.DB.QueryRow(`SELECT id FROM users WHERE username=$1`, "testuser").Scan(&id)
	database.DB.Exec(`UPDATE users SET coins = 0 WHERE id=$1`, id)

	client := &http.Client{}

	item := "pink-hoody"
	url := fmt.Sprint("http://localhost:8080/api/buy/", item)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("error creating request: ", err)
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error making request: ", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestBuyHandlerInvalidItem(t *testing.T) {
	token, err := authUser()
	if err != nil {
		t.Fatal("error auth user: ", err)
	}

	os.Setenv("POSTGRES_URL", localDBURL)
	database.Connect()
	defer database.DB.Close()

	// получаем id, задаём пользователю 0 монет
	var id int
	database.DB.QueryRow(`SELECT id FROM users WHERE username=$1`, "testuser").Scan(&id)
	database.DB.Exec(`UPDATE users SET coins = 0 WHERE id=$1`, id)

	client := &http.Client{}

	item := "hink-poody"
	url := fmt.Sprint("http://localhost:8080/api/buy/", item)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal("error creating request: ", err)
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("error making request: ", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}
}

func authUser() (token string, err error) {
	url := "http://localhost:8080/api/auth"
	userData, err := json.Marshal(map[string]string{"username": "testuser", "password": "tester"})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(userData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var tokenMap map[string]string
	json.NewDecoder(resp.Body).Decode(&tokenMap)
	token = tokenMap["token"]
	if token == "" {
		return "", err
	}

	return token, nil
}
