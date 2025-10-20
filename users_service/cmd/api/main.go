package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"myapp/users_service/internal/handlers"
	"net/http"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=microuser sslmode=disable")
	err = godotenv.Load("users_service/.env")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	usershand := handlers.NewUsersHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/token", usershand.HandlerAuth)
	// mux.HandleFunc("/auth/register", usershand.HandlerRegister)
	log.Fatal(http.ListenAndServe(":8001", mux))
}
