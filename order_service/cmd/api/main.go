package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"myapp/order_service/internal/handlers"
	"net/http"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/microorders?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	orderhandler := handlers.NewOrderHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orders/getall", orderhandler.GetOrders)
	mux.HandleFunc("/v1/orders/create", orderhandler.CreateOrder)
	log.Fatal(http.ListenAndServe(":8002", mux))
}
