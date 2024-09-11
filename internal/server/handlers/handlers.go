package handlers

import (
	"net/http"
)

type DefaultHandler struct{}

func New() DefaultHandler {
	return DefaultHandler{}
}

func (h DefaultHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ok"))
}