package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
		http.Error(w, `{"error": "invalid method"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.Header.Get("X-User-Id")
	if idInt, err := strconv.Atoi(id); err != nil || idInt <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	var totalCount int
	err := o.db.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1", id).Scan(&totalCount)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	var orders []*modules.Order
	query := `
		SELECT * FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	rows, err := o.db.Query(query, id, limit, offset)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var order modules.Order
		err := rows.Scan(&order.Id, &order.UserID, &order.Order,
			&order.Count, &order.Price, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"orders": orders,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_count": totalCount,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

func (o *orderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "invalid method"}`, http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	idStr := query.Get("id")
	if idStr == "" {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}
	iduser := r.Header.Get("X-User-Id")
	if iduser == "" {
		http.Error(w, `{"error": "invalid userid"}`, http.StatusInternalServerError)
		return
	}
	var order modules.Order
	err := o.db.QueryRow(`SELECT * FROM orders WHERE id = $1 and user_id = $2`, idStr, iduser).Scan(
		&order.Id,
		&order.UserID,
		&order.Order,
		&order.Count,
		&order.Price,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (o *orderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "invalid method"}`, http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, `{"error": "invalid body"}`, http.StatusBadRequest)
		return
	}
	id := r.Header.Get("X-User-Id")
	if id == "" {
		http.Error(w, `{"error": "invalid id"}`, http.StatusInternalServerError)
		return
	}
	var order modules.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if order.Id <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}
	query := `UPDATE orders SET `
	args := []interface{}{}
	argscount := 1
	if order.Order != "" {
		query += `"order" = $` + fmt.Sprint(argscount) + `,`
		args = append(args, order.Order)
		argscount++
	}
	if order.Count != 0 {
		query += `count = $` + fmt.Sprint(argscount) + `,`
		args = append(args, order.Count)
		argscount++
	}
	if order.Price != 0 {
		query += `price = $` + fmt.Sprint(argscount) + `,`
		args = append(args, order.Price)
		argscount++
	}
	if order.Status != "" {
		if order.Status == "Новая" || order.Status == "В работе" || order.Status == "Завершен" ||
			order.Status == "Отменен" {
			query += `status = $` + fmt.Sprint(argscount) + `,`
			args = append(args, order.Status)
			argscount++
		} else {
			http.Error(w, `{"error": "invalid status"}`, http.StatusBadRequest)
			return
		}
	}
	query += `updated_at = now() `
	query += `where id = ` + fmt.Sprint(order.Id) + ` and user_id = ` + id + ` returning id, user_id,"order", count, price, status, created_at, updated_at`
	err := o.db.QueryRow(query, args...).Scan(&order.Id, &order.UserID, &order.Order, &order.Count, &order.Price, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "order not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}
