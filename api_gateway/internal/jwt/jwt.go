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
	ID   int      `json:"id"`
	Role []string `json:"roles"`
	jwt.StandardClaims
}

func GenerateToken(id int, roles []string) (string, error) {
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
		ID:   id,
		Role: roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	return token.SignedString([]byte(os.Getenv("SECRETKEY")))
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
