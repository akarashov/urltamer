package handler

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestIsInternalRequest(t *testing.T) {
    tests := []struct {
        name      string
        cidrValue interface{}
        headerIP  string
        wantOK    bool
        wantCode  int
        wantMsg   string
    }{
        {"valid inside", "192.168.0.0/16", "192.168.1.10", true, http.StatusOK, ""},
        {"missing cidr", nil, "192.168.1.10", false, http.StatusForbidden, errMsgForbidden},
        {"invalid cidr", "not-a-cidr", "192.168.1.10", false, http.StatusForbidden, errMsgForbidden},
        {"missing header", "192.168.0.0/16", "", false, http.StatusForbidden, errMsgClientIPMissing},
        {"invalid ip", "192.168.0.0/16", "not-an-ip", false, http.StatusForbidden, errMsgClientIPMissing},
        {"ip not in subnet", "10.0.0.0/8", "192.168.1.10", false, http.StatusForbidden, errMsgClientIPNotInSubnet},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            ctx := context.Background()
            if tc.cidrValue != nil {
                ctx = context.WithValue(ctx, CtxKeyCIDR, tc.cidrValue.(string))
            }
            h := &Handler{Context: ctx}

            req := httptest.NewRequest("GET", "/", nil)
            if tc.headerIP != "" {
                req.Header.Set("X-Real-IP", tc.headerIP)
            }

            ok, code, msg := h.isInternalRequest(req)
            if ok != tc.wantOK {
                t.Fatalf("ok = %v; want %v", ok, tc.wantOK)
            }
            if code != tc.wantCode {
                t.Fatalf("code = %v; want %v", code, tc.wantCode)
            }
            if msg != tc.wantMsg {
                t.Fatalf("msg = %q; want %q", msg, tc.wantMsg)
            }
        })
    }
}
