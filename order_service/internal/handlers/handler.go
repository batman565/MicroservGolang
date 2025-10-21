package handlers

import (
	"database/sql"
	"encoding/json"
	"myapp/order_service/internal/modules"
	"net/http"
	"strconv"
)

type orderHandler struct {
	db *sql.DB
}

func NewOrderHandler(db *sql.DB) *orderHandler {
	return &orderHandler{
		db: db,
	}
}

func (o *orderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "{error: 'invalid method'}", http.StatusMethodNotAllowed)
		return
	}
	id := r.Header.Get("X-User-Id")
	if idInt, err := strconv.Atoi(id); err != nil || idInt <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}
	var orders []*modules.Order
	rows, err := o.db.Query("SELECT * FROM orders WHERE user_id = $1", id)
	if err != nil {
		http.Error(w, "{error: "+err.Error()+"}", http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var order modules.Order
		err := rows.Scan(&order.Id, &order.UserID, &order.Order,
			&order.Count, &order.Price, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			http.Error(w, "{error: "+err.Error()+"}", http.StatusInternalServerError)
			return
		}
		orders = append(orders, &order)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]*modules.Order{"orders": orders})
}

func (o *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "{error: 'invalid method'}", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "{error: 'invalid body'}", http.StatusBadRequest)
		return
	}
	var order modules.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "{error: "+err.Error()+"}", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.Header.Get("X-User-Id"))
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusInternalServerError)
		return
	}
	order.UserID = id
	if order.Price <= 0 || order.Count <= 0 || order.Order == "" {
		http.Error(w, `{"error": "invalid order"}`, http.StatusBadRequest)
		return
	}
	err = o.db.QueryRow(`INSERT INTO orders (user_id, "order", count, price) values ($1, $2, $3, $4) returning id, created_at, updated_at, status`, order.UserID, order.Order, order.Count, order.Price).Scan(
		&order.Id, &order.CreatedAt, &order.UpdatedAt, &order.Status,
	)
	if err != nil {
		http.Error(w, "{error: "+err.Error()+"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}
