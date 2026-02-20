package usecase

import (
	"errors"
	"testing"

	"github.com/kari/brain/internal/domain"
)

type fakeSystemChecker struct {
	report domain.SystemCheckReport
	err    error
	called bool
}

func (f *fakeSystemChecker) RunSystemCheck(_ domain.Server) (domain.SystemCheckReport, error) {
	f.called = true
	return f.report, f.err
}

func TestOnboardRejectsInvalidServer(t *testing.T) {
	checker := &fakeSystemChecker{}
	service := NewServerOnboardingService(checker)

	_, err := service.Onboard(domain.Server{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if checker.called {
		t.Fatal("system check should not run for invalid server")
	}
}

func TestOnboardReturnsSystemReport(t *testing.T) {
	checker := &fakeSystemChecker{
		report: domain.SystemCheckReport{Distro: "ubuntu", Version: "22.04"},
	}
	service := NewServerOnboardingService(checker)

	report, err := service.Onboard(domain.Server{ID: "srv-1", Address: "10.0.0.2"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !checker.called {
		t.Fatal("expected system checker to be called")
	}
	if report.Distro != "ubuntu" {
		t.Fatalf("expected ubuntu distro, got %s", report.Distro)
	}
}
