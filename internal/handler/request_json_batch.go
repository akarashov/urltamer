package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akarashov/urltamer/internal/model"
)

// RequestJSONEndpointBatch handles batch URL shortening requests in JSON format.
func (h *Handler) RequestJSONEndpointBatch(res http.ResponseWriter, req *http.Request) {
	var requestBatchs model.RequestBatchs
	var response model.ResponseBatchs
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
	if err = json.Unmarshal(buf.Bytes(), &requestBatchs); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var batchStatus int = http.StatusCreated
	for _, requestBatch := range requestBatchs {
		uid := generateUserID()
		if id, ok := UserIDFromRequest(req); ok {
			uid = id
		}
		tamer, err := h.Service.CreateShortURL(h.Context, requestBatch.OriginalURL, uid)
		status, ok := h.isConflictResolver(err)
		if !ok {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if status == http.StatusConflict {
			batchStatus = http.StatusConflict
		}
		rsp := model.ResponseBatch{
			CorrelationID: requestBatch.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL)}
		response = append(response, rsp)
	}
	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(res, "Error on marshaling", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(batchStatus)
	res.Write(resp)
}
