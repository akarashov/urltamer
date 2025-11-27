package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func (h *Handler) RequestJSONEndpoint(res http.ResponseWriter, req *http.Request) {
	var request Request
	var response Response
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
	if err = json.Unmarshal(buf.Bytes(), &request); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if request.URL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		uid := generateUserID()
		if cookie, cerr := req.Cookie(COOKIE_NAME); cerr == nil {
			if v := GetUserID(cookie.Value); v > 0 {
				uid = v
			}
		}
		tamer, err := h.Service.CreateShortURL(h.Context, request.URL, uid)
		status, ok := h.isConflictResolver(err)
		if !ok {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		response.Result = fmt.Sprintf("%s/%s", *h.Base, tamer.ShortURL)
		resp, err := json.Marshal(response)
		if err != nil {
			http.Error(res, "Error on marshaling", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(status)
		res.Write(resp)
	}
}
