package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kari/brain/internal/domain"
	"github.com/kari/brain/internal/usecase"
)

type SiteActivationHandler struct {
	service usecase.SiteActivationService
}

type activateSiteRequest struct {
	ID       string `json:"id"`
	Domain   string `json:"domain"`
	IPv4     string `json:"ipv4"`
	IPv6     string `json:"ipv6"`
	OwnerUID int    `json:"owner_uid"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type okResponse struct {
	Status string `json:"status"`
}

var errMethodNotAllowed = errors.New("method not allowed")

func NewSiteActivationHandler(service usecase.SiteActivationService) SiteActivationHandler {
	return SiteActivationHandler{service: service}
}

func (h SiteActivationHandler) HandleActivateSite(w http.ResponseWriter, r *http.Request) {
	req, err := decodeActivateSiteRequest(r)
	if err != nil {
		if errors.Is(err, errMethodNotAllowed) {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err := h.service.Activate(mapSite(req)); err != nil {
		if errors.Is(err, usecase.ErrValidation) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "activation failed"})
		return
	}
	writeJSON(w, http.StatusAccepted, okResponse{Status: "activation_started"})
}

func decodeActivateSiteRequest(r *http.Request) (activateSiteRequest, error) {
	if r.Method != http.MethodPost {
		return activateSiteRequest{}, errMethodNotAllowed
	}
	defer r.Body.Close()

	var req activateSiteRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return activateSiteRequest{}, errors.New("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return activateSiteRequest{}, errors.New("invalid JSON body")
	}
	return req, nil
}

func mapSite(req activateSiteRequest) domain.Site {
	return domain.Site{
		ID:       req.ID,
		Domain:   req.Domain,
		IPv4:     req.IPv4,
		IPv6:     req.IPv6,
		OwnerUID: req.OwnerUID,
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
