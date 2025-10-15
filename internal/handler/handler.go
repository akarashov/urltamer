package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/akarashov/urltamer/internal/config"
)

var (
	mutex  sync.Mutex
	tamers = make(map[string]string)
)

const tamerLength = 8

type requestCfg struct {
	Config *config.Config
}

func makeTamer() string {
	buffer := make([]byte, tamerLength)
	rand.Read(buffer)
	tamer := base64.RawURLEncoding.EncodeToString(buffer)
	return tamer[:tamerLength]
}

func New(c *config.Config) *requestCfg {
	return &requestCfg{
		Config: c,
	}
}

func (rc *requestCfg) RequestEndpoint(res http.ResponseWriter, req *http.Request) {
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
			fmt.Fprintf(res, "%s%s", rc.Config.Base, tamer)
		} else {
			http.Error(res, "Double Tamer", http.StatusBadRequest)
		}
	}
}

func ResponseEndpoint(res http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	tamer := req.URL.Path[1:]
	if reqURL, exist := tamers[tamer]; exist {
		http.Redirect(res, req, reqURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(res, "Not found", http.StatusBadRequest)
	}
}
