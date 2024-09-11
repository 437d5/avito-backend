package server

import (
	"net/http"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/server/handlers/tenders"
	"zadanie-6105/internal/storage/postgres"

	"github.com/gorilla/mux"
)

// LoadRoutes initializes handlers for /tenders/*, /* and /bids/* endpoints
// after initializing it register routes that this handler serves
func LoadRoutes(r *mux.Router, storage *postgres.Storage) {
	defaultHandler := handlers.New()
	r.HandleFunc("/ping", defaultHandler.PingHandler).Methods(http.MethodGet)

	tendersHandler := tenders.New(storage)
	r.HandleFunc("/tenders", tendersHandler.TenderListHandler).Methods(http.MethodGet)
	r.HandleFunc("/tenders/new", tendersHandler.NewTenderHandler).Methods(http.MethodPost)
	r.HandleFunc("/tenders/my", tendersHandler.MyTendersListHandler).Methods(http.MethodGet)
	r.HandleFunc("/tenders/{tenderID}/status", tendersHandler.TenderStatusHandler).Methods(http.MethodGet)
	r.HandleFunc("/tenders/{tenderID}/status", tendersHandler.TenderChangeStatusHandler).Methods(http.MethodPut)
	r.HandleFunc("/tenders/{tenderID}/edit", tendersHandler.EditTenderHandler).Methods(http.MethodPatch)
	r.HandleFunc("/tenders/{tenderID}/rollback/{version}", tendersHandler.RollbackHandler).Methods(http.MethodPut)
}