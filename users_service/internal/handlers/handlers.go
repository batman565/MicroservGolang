package handlers

import (
	"database/sql"
	"encoding/json"
	jwts "myapp/users_service/internal/jwt"
	"myapp/users_service/internal/modules"
	"net/http"
	"slices"
)

type usersHandler struct {
	db *sql.DB
}

func NewUsersHandler(db *sql.DB) *usersHandler {
	return &usersHandler{
		db: db,
	}
}

func (g *usersHandler) HandlerAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	type Auth struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var auth Auth
	err := json.NewDecoder(r.Body).Decode(&auth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if auth.Email == "" || auth.Password == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var UserResponse modules.UserResponse
	var password string
	err = g.db.QueryRow("Select id, password, email, role, created_at, updated_at, name from users where email = $1", auth.Email).Scan(&UserResponse.Id,
		&password,
		&UserResponse.Email,
		&UserResponse.Role,
		&UserResponse.CreatedAt,
		&UserResponse.UpdatedAt,
		&UserResponse.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if jwts.ComparePasswords(password, auth.Password) {
		token, err := jwts.GenerateToken(UserResponse.Id, UserResponse.Role, auth.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		UserResponse.Token = "Bearer " + token
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&UserResponse)
		return
	}
	http.Error(w, "Invalid password", http.StatusUnauthorized)
}

func (g *usersHandler) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	type Register struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Name     string `json:"name"`
	}
	var reg Register
	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if reg.Email == "" || reg.Password == "" || reg.Role == "" ||
		!slices.Contains([]string{"Исполнитель", "Руководитель", "Инженер"}, reg.Role) ||
		reg.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var UserResponse modules.UserResponse
	pass, err := jwts.HashedPassword(reg.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = g.db.QueryRow("insert into users (email, password, name, role) values ($1, $2, $3, $4) returning id, email, name, role, created_at, updated_at", reg.Email, pass, reg.Name, reg.Role).Scan(
		&UserResponse.Id,
		&UserResponse.Email,
		&UserResponse.Name,
		&UserResponse.Role,
		&UserResponse.CreatedAt,
		&UserResponse.UpdatedAt,
	)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	token, err := jwts.GenerateToken(UserResponse.Id, UserResponse.Role, reg.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	UserResponse.Token = "Bearer " + token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&UserResponse)
}
