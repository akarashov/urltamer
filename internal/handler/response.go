package handler

import (
	"net/http"

	"github.com/akarashov/urltamer/internal/service"
)

// ResponseEndpoint handles redirection from a shortened URL to the original URL.
func (h *Handler) ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	tamer := req.URL.Path[1:]
	originalURL, err := h.Service.GetOriginalURL(h.Context, tamer)
	switch err {
	case nil:
		http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
	case service.ErrURLDeleted:
		http.Error(res, "Deleted", http.StatusGone)
	default:
		http.Error(res, "Not found", http.StatusBadRequest)
	}
}
