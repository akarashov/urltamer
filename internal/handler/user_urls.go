package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akarashov/urltamer/internal/model"
)

// UserURLsEndpoint retrieves all shortened URLs for a specific user.
func (h *Handler) UserURLsEndpoint(res http.ResponseWriter, req *http.Request) {
	var userURL model.UserURL
	var userURLs model.UserURLs
	var userID int
	id, ok := UserIDFromRequest(req)
	if ok {
		userID = id
	} else {
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tamers, err := h.Service.GetUserURLs(h.Context, userID)
	if err != nil {
		http.Error(res, "Not found", http.StatusBadRequest)
	} else {
		for _, tamer := range tamers {
			userURL.ShortURL = fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL)
			userURL.OriginalURL = tamer.OriginalURL
			userURLs = append(userURLs, userURL)
		}
		if len(userURLs) == 0 {
			http.Error(res, "No content", http.StatusNoContent)
			return
		}
		resp, err := json.Marshal(userURLs)
		if err != nil {
			http.Error(res, "Error on marshaling", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}
}
