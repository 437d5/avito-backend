package server

import (
	"net/http"
	"zadanie-6105/internal/server/handlers"
	"zadanie-6105/internal/server/handlers/bids"
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

	bidsHandler := bids.New(storage)
	r.HandleFunc("/bids/new", bidsHandler.NewBidHandler).Methods(http.MethodPost)
	r.HandleFunc("/bids/my", bidsHandler.MyBidsListHandler).Methods(http.MethodGet)
	r.HandleFunc("/bids/{tenderID}/list", bidsHandler.GetBidsList).Methods(http.MethodGet)
	r.HandleFunc("/bids/{bidID}/status", bidsHandler.GetBidStatus).Methods(http.MethodGet)
	r.HandleFunc("/bids/{bidID}/status", bidsHandler.ChangeBidStatus).Methods(http.MethodPut)
	r.HandleFunc("/bids/{bidID}/edit", bidsHandler.EditBidHandler).Methods(http.MethodPatch)
	r.HandleFunc("/bids/{bidID}/submit_decision", bidsHandler.SubmitBidHandler).Methods(http.MethodPut)
	r.HandleFunc("/bids/{bidID}/feedback", bidsHandler.SendFeedbackHandler).Methods(http.MethodPut)
	r.HandleFunc("/bids/{bidID}/rollback/{version}", bidsHandler.RollbackHandler).Methods(http.MethodPut)
	r.HandleFunc("/bids/{tenderID}/reviews", bidsHandler.ViewReviewsHandler).Methods(http.MethodGet)
}