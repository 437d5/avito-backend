package bids

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/storage/models"
	"zadanie-6105/internal/storage/postgres"

	"github.com/gorilla/mux"
)

const (
	contentType = "Content-Type"
	appJSON     = "application/json"
)

type BidsHandler struct {
	Storage *postgres.Storage
}

func New(storage *postgres.Storage) *BidsHandler {
	return &BidsHandler{
		Storage: storage,
	}
}

func (h *BidsHandler) NewBidHandler(w http.ResponseWriter, r *http.Request) {
	var newBid models.BidRequest
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&newBid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// FIXME unhandled error
	statusCode, _ := userExists(r, w, h, newBid.AuthorID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	statusCode, _ = isOrganizationResponsible(r, w, h, newBid.AuthorID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	statusCode, _ = tenderExists(r, w, h, newBid.TenderID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	organizationID, err := h.Storage.GetOrganizationID(r.Context(), newBid.AuthorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bid, err := h.Storage.InsertBid(r.Context(), newBid, organizationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bid)
}

func (h *BidsHandler) MyBidsListHandler(w http.ResponseWriter, r *http.Request) {
	statusCode, limit, offset := getLimitAndOfset(w, r)
	if statusCode == HandledError {
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	// FIXME unhandled error
	userID, statusCode, _ := getUserID(w, r, h, username)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	bids, err := h.Storage.GetMyBidsList(r.Context(), limit, offset, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bids)
}

func (h *BidsHandler) GetBidsList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderID"]
	if tenderID == "" {
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

	// FIXME unhandled error
	statusCode, _ := isOrganizationResponsible(r, w, h, userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	statusCode, limit, offset := getLimitAndOfset(w, r)
	if statusCode == HandledError {
		return
	}

	bids, err := h.Storage.GetTenderBids(r.Context(), limit, offset, tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(bids) == 0 {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер или предложение не найдено.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bids)
}

func (h *BidsHandler) GetBidStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidID"]
	if bidID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// FIXME unhandled error
	userID, statusCode, _ := getUserID(w, r, h, username)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	statusCode, _ = isOrganizationResponsible(r, w, h, userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	status, err := h.Storage.GetBidStatus(r.Context(), bidID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if status == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Предложение не найдено.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(status)
}

func (h *BidsHandler) ChangeBidStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidID"]
	if bidID == "" {
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

	// FIXME unhandled error
	statusCode, _ := checkOwner(w, r, h, username, bidID)
	switch statusCode {
	case UnhandledError:
		return
	case HandledError:
		return
	}

	bid, err := h.Storage.ChangeBitStatus(r.Context(), bidID, status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bid)
}

func (h *BidsHandler) EditBidHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	bidID := vars["bidID"]
	log.Println(bidID)
	if bidID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Данные неправильно сформированы или не соответствуют требованиям.", w)
		return
	}

	username := r.URL.Query().Get("username")
	log.Println(username)
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	// FIXME unhandled error
	statusCode, _ := checkOwner(w, r, h, username, bidID)
	log.Println(statusCode)
	switch statusCode {
	case UnhandledError:
		return
	case HandledError:
		return
	}

	var editBid models.EditBidRequest
	err := json.NewDecoder(r.Body).Decode(&editBid)
	log.Println(err)
	if err != nil {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}
	empty := editBid == models.EditBidRequest{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	var namePart, descriptionPart string
	if editBid.Name != "" {
		namePart = fmt.Sprintf("name='%s'", editBid.Name)
	}
	if editBid.Description != "" {
		descriptionPart = fmt.Sprintf("description='%s'", editBid.Description)
	}
	parts := make([]string, 0, 2)

	if namePart != "" {
		parts = append(parts, namePart)
	}
	if descriptionPart != "" {
		parts = append(parts, descriptionPart)
	}
	changeString := strings.Join(parts[:], ", ")
	log.Println(changeString)

	newBid, err := h.Storage.EditBid(r.Context(), bidID, changeString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(newBid)
}

func (h *BidsHandler) SubmitBidHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidID"]
	log.Println(bidID)
	if bidID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Решение не может быть отправлено.", w)
		return
	}

	username := r.URL.Query().Get("username")
	log.Println(username)
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Решение не может быть отправлено.", w)
		return
	}

	decision := r.URL.Query().Get("decision")
	log.Println(decision)
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Решение не может быть отправлено.", w)
		return
	}

	// FIXME unhandled error
	userID, statusCode, _ := getUserID(w, r, h, username)
	log.Println(userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	statusCode, _ = isOrganizationResponsible(r, w, h, userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	log.Println(userOrganizationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	tenderOrganizationID, err := h.Storage.GetBidOrganizationID(r.Context(), bidID)
	log.Println(tenderOrganizationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if tenderOrganizationID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userOrganizationID != tenderOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return
	}

	bid, err := h.Storage.BidDecision(r.Context(), bidID, decision)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	empty := bid == models.Bid{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Предложение не найдено.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bid)
}

func (h *BidsHandler) SendFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidID"]
	log.Println(bidID)
	if bidID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Решение не может быть отправлено.", w)
		return
	}

	username := r.URL.Query().Get("username")
	log.Println(username)
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Решение не может быть отправлено.", w)
		return
	}

	// FIXME unhandled error
	userID, statusCode, _ := getUserID(w, r, h, username)
	log.Println(userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	// FIXME unhandled error
	statusCode, _ = isOrganizationResponsible(r, w, h, userID)
	switch statusCode {
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case HandledError:
		return
	}

	bidFeedback := r.URL.Query().Get("bidFeedback")
	if bidFeedback == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Отзыв не может быть отправлен.", w)
		return
	}

	err := h.Storage.SendFeedback(r.Context(), bidID, userID, bidFeedback)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bid, err := h.Storage.GetBidByID(r.Context(), bidID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	empty := bid == models.Bid{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Предложение не найдено.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bid)
}

func (h *BidsHandler) RollbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["bidID"]
	versionStr := vars["version"]

	if bidID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	if versionStr == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return
	}

	// FIXME unhandled error
	statusCode, _ := checkOwner(w, r, h, username, bidID)
	switch statusCode {
	case UnhandledError:
		return
	case HandledError:
		return
	}

	err = h.Storage.RollbackBid(r.Context(), bidID, version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bid, err := h.Storage.GetBidByID(r.Context(), bidID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	empty := bid == models.Bid{}
	if empty {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Предложение или версия не найдены.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentType, appJSON)
	json.NewEncoder(w).Encode(bid)
}


// FIXME не работаю
func (h *BidsHandler) ViewReviewsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenderID := vars["tenderId"]
	if tenderID == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса: отсутствует идентификатор тендера.", w)
		return
	}
	authorUsername := r.URL.Query().Get("authorUsername")
	if authorUsername == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса: отсутствует имя пользователя автора.", w)
		return
	}
	requesterUsername := r.URL.Query().Get("requesterUsername")
	if requesterUsername == "" {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса: отсутствует имя пользователя запрашивающего.", w)
		return
	}

	statusCode, limit, offset := getLimitAndOfset(w, r)
	if statusCode == HandledError {
		return
	}

	requesterID, statusCode, _ := getUserID(w, r, h, requesterUsername)
	if statusCode == http.StatusNotFound {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь-запрашивающий не найден.", w)
		return
	}

	authorID, statusCode, _ := getUserID(w, r, h, authorUsername)
	if statusCode == http.StatusNotFound {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь-автор не найден.", w)
		return
	}
	statusCode, _ = isOrganizationResponsible(r, w, h, requesterID)
	if statusCode == http.StatusForbidden {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для просмотра отзывов.", w)
		return
	}

	feedback, err := h.Storage.GetFeedback(r.Context(), authorID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(feedback) == 0 {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Отзывы не найдены.", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedback)
}
