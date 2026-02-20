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

type fakeActivationPipeline struct {
	runErr error
}

func (f fakeActivationPipeline) Run(_ domain.Site) error {
	return f.runErr
}

func TestHandleActivateSiteAcceptsValidRequest(t *testing.T) {
	service := usecase.NewSiteActivationService(fakeActivationPipeline{})
	handler := NewSiteActivationHandler(service)

	body := []byte(`{"id":"site-1","domain":"example.com","owner_uid":1001}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sites/activate", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleActivateSite(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d", http.StatusAccepted, res.Code)
	}
}

func TestHandleActivateSiteRejectsInvalidMethod(t *testing.T) {
	service := usecase.NewSiteActivationService(fakeActivationPipeline{})
	handler := NewSiteActivationHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/v1/sites/activate", nil)
	res := httptest.NewRecorder()

	handler.HandleActivateSite(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, res.Code)
	}
}

func TestHandleActivateSiteRejectsUnknownFields(t *testing.T) {
	service := usecase.NewSiteActivationService(fakeActivationPipeline{})
	handler := NewSiteActivationHandler(service)

	body := []byte(`{"id":"site-1","domain":"example.com","owner_uid":1001,"x":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sites/activate", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleActivateSite(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, res.Code)
	}
}

func TestHandleActivateSiteReturnsDomainError(t *testing.T) {
	service := usecase.NewSiteActivationService(fakeActivationPipeline{runErr: errors.New("upstream failure")})
	handler := NewSiteActivationHandler(service)

	body := []byte(`{"id":"site-1","domain":"example.com","owner_uid":1001}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sites/activate", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.HandleActivateSite(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, res.Code)
	}
}
