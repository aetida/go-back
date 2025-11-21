package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	InitDB()

	r := mux.NewRouter()

	// Auth endpoints
	r.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Username, Email, Password string }
		json.NewDecoder(r.Body).Decode(&req)
		user, err := RegisterUser(req.Username, req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		access, refresh, _ := GenerateTokens(user.ID)
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  access,
			"refresh_token": refresh,
			"token_type":    "bearer",
		})
	}).Methods("POST")

	r.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Email, Password string }
		json.NewDecoder(r.Body).Decode(&req)
		user, err := LoginUser(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		access, refresh, _ := GenerateTokens(user.ID)
		json.NewEncoder(w).Encode(map[string]string{
			"access_token":  access,
			"refresh_token": refresh,
			"token_type":    "bearer",
		})
	}).Methods("POST")

	r.HandleFunc("/auth/reset-password", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Email, NewPassword string }
		json.NewDecoder(r.Body).Decode(&req)
		if err := ResetPassword(req.Email, req.NewPassword); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"detail": "Пароль обновлён"})
	}).Methods("POST")

	// WebSocket endpoints
	r.HandleFunc("/create_session", CreateSessionHandler).Methods("GET")
	r.HandleFunc("/ws", WebSocketHandler)

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	http.ListenAndServe(":8000", c.Handler(r))
}
