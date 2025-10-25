package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andre-bernardes200/credsystem-hackathon-2025-10-25/participantes/campeoes-do-canal/configs"
	"github.com/andre-bernardes200/credsystem-hackathon-2025-10-25/participantes/campeoes-do-canal/models"
	"github.com/andre-bernardes200/credsystem-hackathon-2025-10-25/participantes/campeoes-do-canal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		message := map[string]string{"status": "ok"}
		b, _ := json.Marshal(message)
		w.Write(b)
	})
	r.Post("/api/find-service", func(w http.ResponseWriter, r *http.Request) {
		var body models.FindServiceRequest
		if err := models.BindAndValidate(&body, r); err != nil {
			message := models.Message{
				Success: false,
				Data: models.ServiceData{
					ServiceID:   000,
					ServiceName: "000",
				},
				Error: "invalid request body",
			}
			b, _ := json.Marshal(message)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}
		intent, err := service.ClassifyIntent(context.Background(), body.Intent)
		if err != nil {
			message := models.Message{
				Success: false,
				Data: models.ServiceData{
					ServiceID:   000,
					ServiceName: "000",
				},
				Error: "could not classify intent: " + err.Error(),
			}
			b, _ := json.Marshal(message)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(b)
			return
		}
		message := models.Message{
			Success: true,
			Data: models.ServiceData{
				ServiceID:   int(intent.ServiceID),
				ServiceName: intent.ServiceName,
			},
			Error: "",
		}
		b, _ := json.Marshal(message)
		w.Write(b)
	})
	fmt.Printf("\n running on http://localhost%s \n", configs.GetPort())
	http.ListenAndServe(configs.GetPort(), r)
}
