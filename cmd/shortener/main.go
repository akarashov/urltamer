package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sync"
)

var (
	mutex  sync.Mutex
	tamers = make(map[string]string)
)

const tamerLength = 8

func makeTamer() string {
	buffer := make([]byte, tamerLength)
	rand.Read(buffer)
	tamer := base64.RawURLEncoding.EncodeToString(buffer)
	return tamer[:tamerLength]
}

func requestEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	reqURLb, err := io.ReadAll(req.Body)
	reqURL := string(reqURLb)
	if err != nil || reqURL == "" {
		http.Error(res, "Error body parse", http.StatusBadRequest)
	} else {
		tamer := makeTamer()
		if _, exist := tamers[tamer]; !exist {
			for _, v := range tamers {
				if reqURL == v {
					http.Error(res, "Double URL", http.StatusBadRequest)
					return
				}
			}
			tamers[tamer] = reqURL
			res.WriteHeader(http.StatusCreated)
			res.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(res, "http://127.0.0.1:8080/%s\n", tamer)
		} else {
			http.Error(res, "Double Tamer", http.StatusBadRequest)
		}
	}
}

func responseEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	tamer := req.URL.Path[1:]
	if reqURL, exist := tamers[tamer]; exist {
		http.Redirect(res, req, reqURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Not found", http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`GET /`, responseEndpoint)
	mux.HandleFunc(`POST /{$}`, requestEndpoint)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
