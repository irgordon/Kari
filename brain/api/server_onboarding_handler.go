package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kari/brain/internal/domain"
	"github.com/kari/brain/internal/usecase"
)

type ServerOnboardingHandler struct {
	service usecase.ServerOnboardingService
}

type onboardServerRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type onboardServerResponse struct {
	Status string                   `json:"status"`
	Report domain.SystemCheckReport `json:"report"`
}

func NewServerOnboardingHandler(service usecase.ServerOnboardingService) ServerOnboardingHandler {
	return ServerOnboardingHandler{service: service}
}

func (h ServerOnboardingHandler) HandleOnboardServer(w http.ResponseWriter, r *http.Request) {
	req, err := decodeOnboardServerRequest(r)
	if err != nil {
		if errors.Is(err, errMethodNotAllowed) {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	report, err := h.service.Onboard(mapServer(req))
	if err != nil {
		if errors.Is(err, usecase.ErrValidation) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "onboarding failed"})
		return
	}

	writeJSON(w, http.StatusOK, onboardServerResponse{Status: "onboarded", Report: report})
}

func decodeOnboardServerRequest(r *http.Request) (onboardServerRequest, error) {
	if r.Method != http.MethodPost {
		return onboardServerRequest{}, errMethodNotAllowed
	}
	defer r.Body.Close()

	var req onboardServerRequest
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return onboardServerRequest{}, errors.New("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return onboardServerRequest{}, errors.New("invalid JSON body")
	}
	return req, nil
}

func mapServer(req onboardServerRequest) domain.Server {
	return domain.Server{ID: req.ID, Address: req.Address}
}
