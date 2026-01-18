package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/akarashov/urltamer/internal/service"
)

func AuditMiddleware(wrapped http.HandlerFunc, subject service.Subject) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		uid := ""
		action := "shorten"
		id, ok := UserIDFromRequest(req)
		if ok {
			uid = strconv.Itoa(id)
		}
		if req.Method == http.MethodGet {
			action = "follow"
		}

		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyData))
		reqURL := string(bodyData)
		event := model.AuditEvent{
			TS:     int(time.Now().Unix()),
			Action: action,
			UserID: uid,
			URL:    reqURL, // TODO: extract original URL if applicable
		}
		subject.Notify(event)
		wrapped(res, req)
	}
}
