package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akarashov/urltamer/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestResponseEndpoint(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}

	tamers["QAZwsxed"] = "http://example.com"

	tests := []struct {
		name string
		res  http.ResponseWriter
		req  *http.Request
		want want
	}{{
		name: "Valid Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxed", nil),
		want: want{
			statusCode: http.StatusTemporaryRedirect,
			location:   "http://example.com",
		}}, {
		name: "inValid Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxrf", nil),
		want: want{
			statusCode: http.StatusBadRequest,
			location:   "",
		},
	}, {
		name: "Valid Tamer with query",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/QAZwsxed?iddqd=idkfa", nil),
		want: want{
			statusCode: http.StatusTemporaryRedirect,
			location:   "http://example.com",
		},
	}, {
		name: "Blank Tamer",
		res:  httptest.NewRecorder(),
		req:  httptest.NewRequest(http.MethodGet, "/", nil),
		want: want{
			statusCode: http.StatusBadRequest,
			location:   "",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResponseEndpoint(tt.res, tt.req)
			result := tt.res.(*httptest.ResponseRecorder)
			assert.Equal(t, tt.want.location, result.Header().Get("Location"))
			assert.Equal(t, tt.want.statusCode, result.Code)
		})
	}
}

func TestRequestEndpoint(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}

	tamers["QAZwsxed"] = "http://example.com"

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		res  http.ResponseWriter
		req  *http.Request
		want want
	}{
		{
			name: "Valid URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.dev")),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			}}, {
			name: "Blank URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("")),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Error body parse",
			},
		}, {
			name: "Double URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com")),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Double URL",
			},
		}, {
			name: "Long URL",
			res:  httptest.NewRecorder(),
			req:  httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://make-a-url-longer.nathanvarner.com/redirect-to-new-url/index.html?long-url=First%20there%20was%20nothing%20and%20then%20a%20new%20character%20came%20along%20h%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20t%20and%20then%20a%20new%20character%20came%20along%20p%20and%20then%20a%20new%20character%20came%20along%20s%20and%20then%20a%20new%20character%20came%20along%20:%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20/%20and%20then%20a%20new%20character%20came%20along%20y%20and%20then%20a%20new%20character%20came%20along%20a%20and%20then%20a%20new%20character%20came%20along%20.%20and%20then%20a%20new%20character%20came%20along%20r%20and%20then%20a%20new%20character%20came%20along%20u")),
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://127.0.0.1:8080/",
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Config{Base: "http://127.0.0.1:8080/", Listen: ":8080"})
			c.RequestEndpoint(tt.res, tt.req)
			result := tt.res.(*httptest.ResponseRecorder)
			assert.Contains(t, result.Body.String(), tt.want.body)
			assert.Equal(t, tt.want.statusCode, result.Code)
		})
	}
}
