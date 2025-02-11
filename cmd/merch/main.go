package main

import (
	"fmt"
	"net/http"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/database"
	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/routes"
)

func main() {
	database.Connect()
	defer database.DB.Close()
	database.MigrateUP()

	routes.SetupRoutes()

	fmt.Println("Server started.")
	http.ListenAndServe(":8080", nil)
}
