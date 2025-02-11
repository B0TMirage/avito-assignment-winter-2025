package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var err error

	DB, err = sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		fmt.Println(err)
	}
	if err = DB.Ping(); err != nil {
		fmt.Println(err)
	}
}
