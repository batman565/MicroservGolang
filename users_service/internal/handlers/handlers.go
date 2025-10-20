package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	jwts "myapp/users_service/internal/jwt"
	"myapp/users_service/internal/modules"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
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

func (g *usersHandler) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Values("X-User-Id") == nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(r.Header.Values("X-User-Id")[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var User modules.User
	err = g.db.QueryRow("select id, email, name, role, created_at, updated_at from users where id = $1", id).Scan(
		&User.Id,
		&User.Email,
		&User.Name,
		&User.Role,
		&User.CreatedAt,
		&User.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&User)
}

func (g *usersHandler) HandlerGetAllUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []*modules.User
	rows, err := g.db.Query("select id, email, name, role, created_at, updated_at from users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user modules.User
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, &user)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]*modules.User{"users": users})
}

func (g *usersHandler) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		http.Error(w, `{"error": "Request body is required"}`, http.StatusBadRequest)
		return
	}

	userIDHeader := r.Header.Get("X-User-Id")
	if userIDHeader == "" {
		http.Error(w, `{"error": "X-User-Id header is required"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(userIDHeader)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	type userreq struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}

	var user userreq
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid JSON: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if user.Name == "" && user.Email == "" && user.Role == "" && user.Password == "" {
		http.Error(w, `{"error": "At least one field (name, email, role, or password) must be provided"}`, http.StatusBadRequest)
		return
	}

	var args []interface{}
	var setClauses []string
	paramCount := 1

	if user.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramCount))
		args = append(args, user.Name)
		paramCount++
	}

	if user.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", paramCount))
		args = append(args, user.Email)
		paramCount++
	}

	if user.Role != "" {
		validRoles := []string{"Исполнитель", "Руководитель", "Инженер"}
		if !slices.Contains(validRoles, user.Role) {
			http.Error(w, `{"error": "Invalid role. Must be one of: Исполнитель, Руководитель, Инженер"}`, http.StatusBadRequest)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", paramCount))
		args = append(args, user.Role)
		paramCount++
	}

	if user.Password != "" {
		hashedPassword, err := jwts.HashedPassword(user.Password)
		if err != nil {
			http.Error(w, `{"error": "Failed to hash password: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", paramCount))
		args = append(args, hashedPassword)
		paramCount++
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramCount))
	args = append(args, time.Now())
	paramCount++

	args = append(args, id)

	query := "UPDATE users SET " + strings.Join(setClauses, ", ") +
		fmt.Sprintf(" WHERE id = $%d", paramCount)

	result, err := g.db.Exec(query, args...)
	if err != nil {
		http.Error(w, `{"error": "Database error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, `{"error": "Failed to check affected rows: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]any{
		"message": "User updated successfully",
		"user_id": id,
		"updated_fields": map[string]any{
			"name":     user.Name != "",
			"email":    user.Email != "",
			"role":     user.Role != "",
			"password": user.Password != "",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (g *usersHandler) HandlerUpdateUserAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		http.Error(w, `{"error": "Request body is required"}`, http.StatusBadRequest)
		return
	}

	type userreq struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}

	var user userreq
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid JSON: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if user.Id == 0 || (user.Name == "" && user.Email == "" && user.Role == "" && user.Password == "") {
		http.Error(w, `{"error": "At least one field (name, email, role, or password) and id must be provided"}`, http.StatusBadRequest)
		return
	}

	var args []interface{}
	var setClauses []string
	paramCount := 1

	if user.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramCount))
		args = append(args, user.Name)
		paramCount++
	}

	if user.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", paramCount))
		args = append(args, user.Email)
		paramCount++
	}

	if user.Role != "" {
		validRoles := []string{"Исполнитель", "Руководитель", "Инженер"}
		if !slices.Contains(validRoles, user.Role) {
			http.Error(w, `{"error": "Invalid role. Must be one of: Исполнитель, Руководитель, Инженер"}`, http.StatusBadRequest)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", paramCount))
		args = append(args, user.Role)
		paramCount++
	}

	if user.Password != "" {
		hashedPassword, err := jwts.HashedPassword(user.Password)
		if err != nil {
			http.Error(w, `{"error": "Failed to hash password: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", paramCount))
		args = append(args, hashedPassword)
		paramCount++
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramCount))
	args = append(args, time.Now())
	paramCount++

	args = append(args, user.Id)

	query := "UPDATE users SET " + strings.Join(setClauses, ", ") +
		fmt.Sprintf(" WHERE id = $%d", paramCount)

	result, err := g.db.Exec(query, args...)
	if err != nil {
		http.Error(w, `{"error": "Database error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, `{"error": "Failed to check affected rows: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]any{
		"message": "User updated successfully",
		"user_id": user.Id,
		"updated_fields": map[string]any{
			"name":     user.Name != "",
			"email":    user.Email != "",
			"role":     user.Role != "",
			"password": user.Password != "",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
