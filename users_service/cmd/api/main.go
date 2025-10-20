package main

import (
	"database/sql"
	"log"
	"myapp/users_service/internal/handlers"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=microuser sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load("users_service/.env")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	usershand := handlers.NewUsersHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", usershand.HandlerAuth)
	mux.HandleFunc("/v1/auth/register", usershand.HandlerRegister)
	mux.HandleFunc("/v1/users/get", usershand.HandlerGetUser)
	log.Fatal(http.ListenAndServe(":8001", mux))
}
