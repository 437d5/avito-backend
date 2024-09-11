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
	tendersListParams := []string{"limit", "{limit}", "offset", "{offset}", "service_type", "{service_type}"}
	r.HandleFunc("/tenders", tendersHandler.TenderListHandler).Queries(tendersListParams...).Methods(http.MethodGet)
	r.HandleFunc("/tenders/new", tendersHandler.NewTenderHandler).Methods(http.MethodPost)
	myTendersListParams := []string{"limit", "{limit}", "offset", "{offset}", "username", "{username}"}
	r.HandleFunc("/tenders/my", tendersHandler.MyTendersListHandler).Queries(myTendersListParams...).Methods(http.MethodGet)
}