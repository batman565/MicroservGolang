package handlers

import "database/sql"

type orderHandler struct {
	db *sql.DB
}

func NewOrderHandler(db *sql.DB) *orderHandler {
	return &orderHandler{
		db: db,
	}
}

func (o *orderHandler) GetOrder() {}
