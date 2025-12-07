package handler

import (
	"bytes"
	"encoding/json"

	"net/http"
)

func (h *Handler) DeleteUserURLsEndpoint(res http.ResponseWriter, req *http.Request) {
	var shortURLs []string
	var buf bytes.Buffer

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &shortURLs); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(shortURLs) == 0 {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		id, ok := UserIDFromRequest(req)
		if ok {
			uid := id
		affected, err := h.Service.DeleteTamer(h.Context, uid, shortURLs)
		if err == nil && affected >0 {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusAccepted)
			return
		} else if err == nil && affected == 0 {
			http.Error(res, "Not found", http.StatusNotFound)
			return
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		}
	}
}
