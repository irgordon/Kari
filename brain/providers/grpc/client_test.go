package grpc

import (
	"errors"
	"testing"

	"github.com/kari/brain/internal/domain"
)

type fakeTransport struct {
	lastSystemCheckRequest RunSystemCheckRequest
	lastActivateRequest    ActivateSiteRequest
	systemCheckResponse    RunSystemCheckResponse
	activateResponse       ActivateSiteResponse
	systemCheckErr         error
	activateErr            error
}

func (f *fakeTransport) RunSystemCheck(request RunSystemCheckRequest) (RunSystemCheckResponse, error) {
	f.lastSystemCheckRequest = request
	return f.systemCheckResponse, f.systemCheckErr
}

func (f *fakeTransport) ActivateSite(request ActivateSiteRequest) (ActivateSiteResponse, error) {
	f.lastActivateRequest = request
	return f.activateResponse, f.activateErr
}

func TestRunSystemCheckMapsTransportResponse(t *testing.T) {
	transport := &fakeTransport{systemCheckResponse: RunSystemCheckResponse{Distro: "ubuntu", Version: "22.04"}}
	client := NewClientWithTransport(transport)

	report, err := client.RunSystemCheck(domain.Server{ID: "srv-1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if transport.lastSystemCheckRequest.ServerID != "srv-1" {
		t.Fatalf("expected server id srv-1, got %s", transport.lastSystemCheckRequest.ServerID)
	}
	if report.Distro != "ubuntu" {
		t.Fatalf("expected distro ubuntu, got %s", report.Distro)
	}
}

func TestCreateSystemUserSendsAction(t *testing.T) {
	transport := &fakeTransport{activateResponse: ActivateSiteResponse{OK: true}}
	client := NewClientWithTransport(transport)

	err := client.CreateSystemUser(domain.Site{ID: "site-1", Domain: "example.com", OwnerUID: 1001})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if transport.lastActivateRequest.Action != ActionCreateSystemUser {
		t.Fatalf("expected %s action, got %s", ActionCreateSystemUser, transport.lastActivateRequest.Action)
	}
}

func TestActivateSiteReturnsTransportError(t *testing.T) {
	transport := &fakeTransport{activateErr: errors.New("transport failed")}
	client := NewClientWithTransport(transport)

	err := client.ApplyHTTPVHost(domain.Site{ID: "site-1", Domain: "example.com", OwnerUID: 1001})
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestActivateSiteReturnsRejectionWhenNotOK(t *testing.T) {
	transport := &fakeTransport{activateResponse: ActivateSiteResponse{OK: false}}
	client := NewClientWithTransport(transport)

	err := client.IssueCertificate(domain.Site{ID: "site-1", Domain: "example.com", OwnerUID: 1001})
	if !errors.Is(err, errActionRejected) {
		t.Fatalf("expected errActionRejected, got %v", err)
	}
}
