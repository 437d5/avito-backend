package tenders

import (
	"encoding/json"
	"net/http"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/storage/postgres"
)

type TendersHandler struct {
	Storage *postgres.Storage
}

func New(storage *postgres.Storage) *TendersHandler {
	return &TendersHandler{
		Storage: storage,
	}
}

// Список тендеров с возможностью фильтрации по типу услуг.
// Если фильтры не заданы, возвращаются все тендеры.
func (h *TendersHandler) TenderListHandler(w http.ResponseWriter, r *http.Request) {
	limit, err := handlers.ParseQueryParam(r, "limit", 5)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(handlers.ErrorResponse{
			Reason: "Неверный формат запроса или его параметры.",
		})
		return
	}

	offset, err := handlers.ParseQueryParam(r, "offset", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(handlers.ErrorResponse{
			Reason: "Неверный формат запроса или его параметры.",
		})
		return
	}

	serviceType := r.URL.Query().Get("service_type")

	tendersList, err := h.Storage.GetTenderList(r.Context(), limit, offset, serviceType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if tendersList == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tendersList)
}

// Создание нового тендера с заданными параметрами.
// TODO написать реализацию метода
func (h *TendersHandler) NewTenderHandler(w http.ResponseWriter, r *http.Request) {}

// Получение списка тендеров текущего пользователя.
// Для удобства использования включена поддержка пагинации.
// TODO написать реализацию метода
func (h *TendersHandler) MyTendersListHandler(w http.ResponseWriter, r *http.Request) {}

