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
