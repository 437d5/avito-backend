package handlers

import (
	"encoding/json"
	"net/http"
	"zadanie-6105/internal/storage/postgres"
)

type Handler struct {
	Storage *postgres.Storage
}

type errorResponse struct {
	Reason string `json:"reason"`
}

func New(storage *postgres.Storage) Handler {
	return Handler{
		Storage: storage,
	}
}

func (h Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Reason: "Неверный метод запроса",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}