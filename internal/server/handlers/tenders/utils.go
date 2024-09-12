package tenders

import (
	"fmt"
	"net/http"
	"zadanie-6105/internal/server/handlers"
)

const (
	Authorized = 0
	HandledError = 1
	UnhandledError = 2
)

func validateUserAndOrganization(r *http.Request, w http.ResponseWriter, h *TendersHandler, username, tenderID string) (int, error) {
	userID, err := h.Storage.GetUserID(r.Context(), username)
	if err != nil {
		return UnhandledError, fmt.Errorf("failed to get user ID: %w", err)
	}
	if userID == "" {
		handlers.ReturnErrorResponse(http.StatusUnauthorized, "Пользователь не существует или некорректен.", w)
		return HandledError, nil
	}

	userOrganizationID, err := h.Storage.GetOrganizationID(r.Context(), userID)
	if err != nil {
		return UnhandledError, fmt.Errorf("failed to get organization ID: %w", err)
	}
	if userOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return HandledError, nil
	}

	tenderOrganizationID, err := h.Storage.GetOrganizationIDByTender(r.Context(), tenderID)
	if err != nil {
		return UnhandledError, fmt.Errorf("failed to get organization ID: %w", err)
	}
	if tenderOrganizationID == "" {
		handlers.ReturnErrorResponse(http.StatusNotFound, "Тендер не найден.", w)
		return HandledError, nil
	}

	if tenderOrganizationID != userOrganizationID {
		handlers.ReturnErrorResponse(http.StatusForbidden, "Недостаточно прав для выполнения действия.", w)
		return HandledError, nil
	}

	return Authorized, nil
}