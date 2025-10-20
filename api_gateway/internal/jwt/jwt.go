package jwts

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleDirector     string = "Руководитель"
	RoleManager      string = "Менеджер"
	RoleEngineer     string = "Инженер"
	claimsContextKey string = "user"
)

type Claims struct {
	ID    int    `json:"id"`
	Role  string `json:"roles"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("token is invalid")
}

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func ComparePasswords(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		token_string := parts[1]
		claims, err := ValidateToken(token_string)
		if err != nil {
			http.Error(w, `{"error": "Invalid token`+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	if !ok {
		return nil, fmt.Errorf("claims not found in context")
	}
	return claims, nil
}

func RequireRole(allowed_role ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaimsFromContext(r.Context())
			if err != nil {
				http.Error(w, `{"error": "Unauthorized`+err.Error()+`"}`, http.StatusUnauthorized)
				return
			}
			has := false
			for _, role := range allowed_role {
				if claims.Role == role {
					has = true
				}
			}
			if !has {
				http.Error(w, `{"error": "Access denied."}`, http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}

func RequireManager(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleManager)(next))
}

func RequireEngineer(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleEngineer)(next))
}

func RequireSupervisor(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleDirector)(next))
}

func RequireManagerOrEngineerOrSupervisor(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleDirector, RoleManager, RoleEngineer)(next))
}

func RequireSupervisorOrManager(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleDirector, RoleManager)(next))
}

func RequireManagerOrEngineer(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(RequireRole(RoleManager, RoleEngineer)(next))
}
