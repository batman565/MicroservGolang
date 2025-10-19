package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

type Claims struct {
	ID    int      `json:"id"`
	Role  []string `json:"roles"`
	Email string   `json:"email"`
	jwt.StandardClaims
}

func GenerateToken(id int, roles []string, email string) (string, error) {
	err := godotenv.Load("api_gateway/.env")
	if err != nil {
		return "", err
	}
	if len(roles) > 0 {
		return "", fmt.Errorf("roles not implemented")
	}
	if id <= 0 {
		return "", fmt.Errorf("id must be greater more than 0")
	}
	Claims := &Claims{
		ID:    id,
		Role:  roles,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
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
