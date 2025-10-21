package main

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	jwts "myapp/api_gateway/internal/jwt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	godotenv.Load("api_gateway/.env")
	userUrl, err := url.Parse(os.Getenv("USER_SERVICE"))
	if err != nil {
		log.Fatal(err)
	}
	userProxy := httputil.NewSingleHostReverseProxy(userUrl)

	orderUrl, err := url.Parse(os.Getenv("ORDER_SERVICE"))
	if err != nil {
		log.Fatal(err)
	}
	orderProxy := httputil.NewSingleHostReverseProxy(orderUrl)
	mux := http.NewServeMux()
	//Users
	mux.Handle("/v1/auth/token", createProxyHandler(userProxy))
	mux.Handle("/v1/auth/register", createProxyHandler(userProxy))
	mux.Handle("/v1/users/get", jwts.AuthMiddleware(createProxyHandler(userProxy)))
	mux.Handle("/v1/users/getall", jwts.RequireSupervisorOrManager(createProxyHandler(userProxy)))
	mux.Handle("/v1/users/update", jwts.AuthMiddleware(createProxyHandler(userProxy)))
	mux.Handle("/v1/users/update/admin", jwts.RequireSupervisorOrManager(createProxyHandler(userProxy)))
	//Orders
	mux.Handle("/v1/orders/getall", jwts.AuthMiddleware(createProxyHandler(orderProxy)))
	mux.Handle("/v1/orders/create", jwts.AuthMiddleware(createProxyHandler(orderProxy)))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func createProxyHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if claims, err := jwts.GetClaimsFromContext(r.Context()); err == nil {
			r.Header.Set("X-User-Id", fmt.Sprint(claims.ID))
			r.Header.Set("X-User-Role", claims.Role)
			r.Header.Set("X-User-Email", claims.Email)
		}
		log.Printf("Проксирование  на %s", r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}
