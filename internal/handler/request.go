package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/akarashov/urltamer/internal/service"
)

func (h *Handler) isConflictResolver(err error) (int, bool) {
	if err == nil {
		return http.StatusCreated, true
	} else if errors.Is(err, service.ErrURLAlreadyExists) {
		return http.StatusConflict, true
	} else {
		return 0, false
	}
}

func (h *Handler) RequestEndpoint(res http.ResponseWriter, req *http.Request) {
	reqURLb, err := io.ReadAll(req.Body)
	reqURL := string(reqURLb)
	if err != nil || reqURL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		uid := generateUserID()
		if id, ok := UserIDFromRequest(req); ok {
			uid = id
		}
		tamer, err := h.Service.CreateShortURL(h.Context, reqURL, uid)
		status, ok := h.isConflictResolver(err)
		if !ok {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(status)
		resp := []byte(fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL))
		res.Write(resp)
	}
}
