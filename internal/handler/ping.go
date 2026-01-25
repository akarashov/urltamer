package handler

import (
	"net/http"
)

// PingEndpoint checks the connectivity to the database.
func (h *Handler) PingEndpoint(res http.ResponseWriter, req *http.Request) {
	if h.Service.Ping(h.Context) {
		res.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(res, "Not connected to DB", http.StatusInternalServerError)
	}
}
