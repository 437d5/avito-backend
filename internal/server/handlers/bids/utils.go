package bids

import (
	"fmt"
	"net/http"
	"zadanie-6105/internal/server/handlers"
)

const (
	Ok       = 0
	HandledError   = 1
	UnhandledError = 2
)

// func validateUserAndOrganization() (int, error) {}

func userExists(r *http.Request, w http.ResponseWriter, h *BidsHandler, userID string) (int, error) {
	ok, err := h.Storage.UserExists(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, fmt.Errorf("cannot authorize user: %w", err)
	}
	if !ok {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return HandledError, nil
	}

	return Ok, nil
}

func isOrganizationResponsible(r *http.Request, w http.ResponseWriter, h *BidsHandler, userID string) (int, error) {
	organizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, fmt.Errorf("cannot authorize user: %w", err)
	}
	if organizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return HandledError, nil
	}

	return Ok, nil
}

func tenderExists(r *http.Request, w http.ResponseWriter, h *BidsHandler, tenderID string) (int, error) {
	ok, err := h.Storage.TenderExists(r.Context(), tenderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, fmt.Errorf("cannot find tender: %w", err)
	}
	if !ok {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return HandledError, nil
	}

	return Ok, nil
}

func getLimitAndOfset(w http.ResponseWriter, r *http.Request) (int, int, int) {
	limit, err := handlers.ParseQueryParam(r, "limit", 5)
	if err != nil {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return HandledError, limit, 0
	}

	offset, err := handlers.ParseQueryParam(r, "offset", 0)
	if err != nil {
		handlers.ReturnErrorResponse(http.StatusBadRequest, "Неверный формат запроса или его параметры.", w)
		return HandledError, limit, offset
	}

	return Ok, limit, offset
}

func getUserID(w http.ResponseWriter, r *http.Request, h *BidsHandler, username string) (string, int, error) {
	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", UnhandledError, fmt.Errorf("cannot get userID: %w", err)
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return  "", HandledError, nil
	}

	return userID, Ok, nil
}

func checkOwner(w http.ResponseWriter, r *http.Request, h *BidsHandler, username, bidID string) (int, error) {
	userID, statusCode, err := getUserID(w, r, h, username)
	switch statusCode {
	case HandledError:
		return HandledError, nil
	case UnhandledError:
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, fmt.Errorf("cannot check userID: %w", err)
	}

	organizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, err
	}
	if organizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return HandledError, err
	}

	bidOrganizationID, err := h.Storage.GetBidOrganizationID(r.Context(), bidID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return UnhandledError, fmt.Errorf("cannot check organizationID: %w", err)
	}
	if bidOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Предложение не найдено.", w)
		return HandledError, nil
	}

	if organizationID != bidOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return HandledError, nil
	}

	return Ok, nil
}