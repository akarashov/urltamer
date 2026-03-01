package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/akarashov/urltamer/internal/model"
)

type CtxKey string

const CtxKeyCIDR CtxKey = "CIDR"

const (
	errMsgForbidden           = "Forbidden"
	errMsgClientIPMissing     = "Client IP is missing"
	errMsgClientIPNotInSubnet = "Client IP is not a part of subnet"
)

// UserURLsEndpoint retrieves all shortened URLs for a specific user.
func (h *Handler) InternalStatsEndpoint(res http.ResponseWriter, req *http.Request) {
	var internalStats *model.InternalStats

	ok, status, msg := h.isInternalRequest(req)
	if !ok {
		http.Error(res, msg, status)
		return
	}

	internalStats, err := h.Service.GetInternalStats(h.Context)
	if err != nil {
		log.Printf("InternalStatsEndpoint: service error: %v", err)
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(internalStats); err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// isInternalRequest checks if the incoming request is from an internal source
// based on the client's IP address and a trusted subnet defined in the configuration.
func (h *Handler) isInternalRequest(req *http.Request) (bool, int, string) {
	cidr, ok := h.Context.Value(CtxKeyCIDR).(string)
	if !ok || cidr == "" {
		return false, http.StatusForbidden, errMsgForbidden
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Printf("isInternalRequest: invalid CIDR in context: %v", err)
		return false, http.StatusForbidden, errMsgForbidden
	}

	realIP := req.Header.Get("X-Real-IP")
	if realIP == "" {
		return false, http.StatusForbidden, errMsgClientIPMissing
	}
	ip := net.ParseIP(realIP)
	if ip == nil {
		return false, http.StatusForbidden, errMsgClientIPMissing
	}
	if !ipnet.Contains(ip) {
		return false, http.StatusForbidden, errMsgClientIPNotInSubnet
	}
	return true, http.StatusOK, ""
}
