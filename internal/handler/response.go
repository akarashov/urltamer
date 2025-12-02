package handler

import (
	"net/http"
)

func (h *Handler) ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	tamer := req.URL.Path[1:]
	originalURL, err := h.Service.GetOriginalURL(h.Context, tamer)
	if err != nil {
		http.Error(res, "Not found", http.StatusBadRequest)
	} else if originalURL == "#" {
		http.Error(res, "Deleted", http.StatusGone)
	} else {
		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	}
}
