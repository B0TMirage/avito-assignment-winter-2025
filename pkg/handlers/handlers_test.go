package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"

	_ "github.com/lib/pq"
)

func TestAuthHandler(t *testing.T) {
	url := "http://localhost:8080/api/auth"
	tests := []struct {
		name         string
		userData     map[string]string
		wantedStatus int
	}{
		{
			name:         "Correct",
			userData:     map[string]string{"username": "testuser", "password": "tester"},
			wantedStatus: http.StatusOK,
		},
		{
			name:         "Null password",
			userData:     map[string]string{"username": "testuserBadPassword", "password": ""},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "Password length is less than 4",
			userData:     map[string]string{"username": "testuserBadPassword", "password": "404"},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name:         "Null username",
			userData:     map[string]string{"username": "", "password": "avito"},
			wantedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.userData)
			if err != nil {
				t.Error("error marshaling JSON:", err)
				return
			}

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Error("error making request:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantedStatus {
				t.Errorf("got status %v, want %v", resp.StatusCode, tt.wantedStatus)
			}

			if resp.StatusCode == http.StatusOK {
				var token map[string]string
				json.NewDecoder(resp.Body).Decode(&token)
				if token["token"] == "" {
					t.Error(`got "", want token`)
				}
				database.Connect()
				defer database.DB.Close()
				err := database.DB.QueryRow("SELECT id FROM users WHERE username=$1", tt.userData["username"]).Err()
				if err == sql.ErrNoRows {
					t.Error("user was not created")
				}
			}
		})
	}
}

func TestGetAllInfoHandler(t *testing.T) {
	url := "http://localhost:8080/api/auth"
	userdata := map[string]string{"username": "testusers", "password": "tester"}
	jsonData, err := json.Marshal(userdata)
	if err != nil {
		t.Error("error marshaling JSON:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Error("error making request:", err)
		return
	}
	defer resp.Body.Close()

	tokenMap := map[string]string{}
	json.NewDecoder(resp.Body).Decode(&tokenMap)

	token := tokenMap["token"]

	url = "http://localhost:8080/api/info"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error("error creating request:", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Error("error making request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("failed to get resource: %s\n", resp.Status)
	}
}
