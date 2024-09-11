package tenders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/storage/models"
	"zadanie-6105/internal/storage/postgres"

	"github.com/gorilla/mux"
)

const (
	contentType = "Content-Type"
	appJSON = "application/json"
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
		return
	}

	offset, err := handlers.ParseQueryParam(r, "offset", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tendersList)
}

// Создание нового тендера с заданными параметрами.
func (h *TendersHandler) NewTenderHandler(w http.ResponseWriter, r *http.Request) {
	var newTender models.NewTenderRequest

	err := json.NewDecoder(r.Body).Decode(&newTender)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	username := newTender.CreatorUsername
	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return
	}

	organizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if organizationID == "" || newTender.OrganizationID != organizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tender, err := h.Storage.InsertTender(r.Context(), &newTender, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tender)
}

// Получение списка тендеров текущего пользователя.
// Для удобства использования включена поддержка пагинации.
func (h *TendersHandler) MyTendersListHandler(w http.ResponseWriter, r *http.Request) {
	limit, err := handlers.ParseQueryParam(r, "limit", 5)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	offset, err := handlers.ParseQueryParam(r, "offset", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil || userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
	}

	tenders, err := h.Storage.GetMyTendersList(r.Context(), limit, offset, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tenders)
}

// Получить статус тендера по его уникальному идентификатору.
// 401 Если пользователь не существует или некорректен
// 403 Если у пользователя нету прав на выполнение
// 404 Если тендер с таким id не найден
func (h *TendersHandler) TenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderID"]
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return
	}

	// FIXME не возвращает tenderOrganizationID
	tenderOrganizationID, err := h.Storage.GetOrganizationIDByTender(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tenderOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}
	
	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	if tenderOrganizationID != userOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	status, err := h.Storage.GetTenderStatus(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if status == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}
	
	switch status {
	case "CREATED":
		status = "Created"
	case "PUBLISHED":
		status = "Published"
	case "CLOSED":
		status = "Closed" 
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(status)
}

// Изменить статус тендера по его идентификатору.
func (h *TendersHandler) TenderChangeStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderID"]
	if tenderID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return
	}

	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tenderOrganizationID, err := h.Storage.GetOrganizationIDByTender(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tenderOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}

	tender, err := h.Storage.ChangeTenderStatus(r.Context(), tenderID, status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	empty := tender == models.Tender{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tender)
}

// FIXME empty body crash 500
// Изменение параметров существующего тендера.
func (h *TendersHandler) EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderID"]
	if tenderID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Данные неправильно сформированы или не соответствуют требованиям.", w)
		return
	}
	
	username := r.URL.Query().Get("username")
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Данные неправильно сформированы или не соответствуют требованиям.", w)
		return
	}

	var editTender models.EditTenderRequest
	err := json.NewDecoder(r.Body).Decode(&editTender)
	if err != nil {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Данные неправильно сформированы или не соответствуют требованиям.", w)
		return
	}

	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return
	}

	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tenderOrganizationID, err := h.Storage.GetOrganizationIDByTender(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tenderOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}

	if tenderOrganizationID != userOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	var namePart, descriptionPart, serviceTypePart string
	if editTender.Name != "" {
		namePart = fmt.Sprintf("name='%s'", editTender.Name)
	}
	if editTender.Description != "" {
		descriptionPart = fmt.Sprintf("description='%s'", editTender.Description)
	}
	if editTender.ServiceType != "" {
		serviceTypePart = fmt.Sprintf("service_type='%s'", editTender.ServiceType)
	}
	parts := make([]string, 0, 3)

	if namePart != "" {
		parts = append(parts, namePart)
	}
	if descriptionPart != "" {
		parts = append(parts, descriptionPart)
	}
	if serviceTypePart != "" {
		parts = append(parts, serviceTypePart)
	}
	changeString := strings.Join(parts[:], ", ")

	tender, err := h.Storage.EditTender(r.Context(), tenderID, changeString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	empty := tender == models.Tender{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tender)
}

// FIXME not found not working
func (h *TendersHandler) RollbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderID"]
	version := vars["version"]

	if tenderID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}
	if version == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return
	}

	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tenderOrganizationID, err := h.Storage.GetOrganizationIDByTender(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер или версия не найдены.", w)
		return
	}

	if userOrganizationID != tenderOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tender, err := h.Storage.RollbackTender(r.Context(), tenderID, version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	empty := tender == models.Tender{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(tender)
}