package usecase

import (
	"errors"
	"testing"

	"github.com/kari/brain/internal/domain"
)

type fakePipeline struct {
	runErr error
	ran    bool
}

func (f *fakePipeline) Run(_ domain.Site) error {
	f.ran = true
	return f.runErr
}

func TestActivateRejectsInvalidSite(t *testing.T) {
	pipeline := &fakePipeline{}
	service := NewSiteActivationService(pipeline)

	err := service.Activate(domain.Site{ID: "", Domain: "", OwnerUID: 0})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if pipeline.ran {
		t.Fatal("pipeline should not run for invalid site")
	}
}

func TestActivateRunsPipelineForValidSite(t *testing.T) {
	pipeline := &fakePipeline{}
	service := NewSiteActivationService(pipeline)

	err := service.Activate(domain.Site{ID: "site-1", Domain: "example.com", OwnerUID: 1001})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !pipeline.ran {
		t.Fatal("pipeline should run for valid site")
	}
}

func TestActivateReturnsPipelineError(t *testing.T) {
	pipeline := &fakePipeline{runErr: errors.New("pipeline failed")}
	service := NewSiteActivationService(pipeline)

	err := service.Activate(domain.Site{ID: "site-1", Domain: "example.com", OwnerUID: 1001})
	if err == nil {
		t.Fatal("expected pipeline error")
	}
}
