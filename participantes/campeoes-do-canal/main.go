package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andre-bernardes200/credsystem-hackathon-2025-10-25/participantes/campeoes-do-canal/configs"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		message := map[string]string{"status": "ok"}
		b, _ := json.Marshal(message)
		w.Write(b)
	})
	r.Post("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		type FindServiceRequest struct {
			Intent string `json:"intent" validate:"required"`
		}
		var body FindServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		w.Write([]byte("ok"))
	})
	fmt.Printf("\n running on http://localhost%s \n", configs.GetPort())
	http.ListenAndServe(configs.GetPort(), r)
}
