package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kari/brain/internal/domain"
	"github.com/kari/brain/internal/usecase"
)

type fakeSystemCheckForHandler struct {
	report domain.SystemCheckReport
	err    error
}

func (f fakeSystemCheckForHandler) RunSystemCheck(_ domain.Server) (domain.SystemCheckReport, error) {
	return f.report, f.err
}

func TestHandleOnboardServerAcceptsValidRequest(t *testing.T) {
	service := usecase.NewServerOnboardingService(fakeSystemCheckForHandler{report: domain.SystemCheckReport{Distro: "ubuntu"}})
	handler := NewServerOnboardingHandler(service)

	body := []byte(`{"id":"srv-1","address":"10.0.0.2"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/servers/onboard", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleOnboardServer(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, res.Code)
	}
}

func TestHandleOnboardServerRejectsInvalidMethod(t *testing.T) {
	service := usecase.NewServerOnboardingService(fakeSystemCheckForHandler{})
	handler := NewServerOnboardingHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/v1/servers/onboard", nil)
	res := httptest.NewRecorder()

	handler.HandleOnboardServer(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, res.Code)
	}
}

func TestHandleOnboardServerRejectsInvalidJSON(t *testing.T) {
	service := usecase.NewServerOnboardingService(fakeSystemCheckForHandler{})
	handler := NewServerOnboardingHandler(service)

	body := []byte(`{"id":"srv-1","address":"10.0.0.2","x":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/servers/onboard", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleOnboardServer(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, res.Code)
	}
}

func TestHandleOnboardServerReturnsExecutionFailure(t *testing.T) {
	service := usecase.NewServerOnboardingService(fakeSystemCheckForHandler{err: errors.New("agent offline")})
	handler := NewServerOnboardingHandler(service)

	body := []byte(`{"id":"srv-1","address":"10.0.0.2"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/servers/onboard", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleOnboardServer(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, res.Code)
	}
}
