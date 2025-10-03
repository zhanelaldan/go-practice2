package main

import (
	"fmt"
	"log"
	"net/http"

	"go-practice2/internal/handlers"
	"go-practice2/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	// маршруты
	mux.Handle("/user", middleware.WithAPIKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetUser(w, r)
		} else if r.Method == http.MethodPost {
			handlers.PostUser(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
