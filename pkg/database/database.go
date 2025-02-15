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

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(50)
	DB.SetConnMaxLifetime(0)

	if err = DB.Ping(); err != nil {
		fmt.Println(err)
	}
}
