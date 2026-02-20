package usecase

import (
	"errors"
	"strings"

	"github.com/kari/brain/internal/domain"
)

type SiteActivationPipeline interface {
	Run(site domain.Site) error
}

type SiteActivationService struct {
	pipeline SiteActivationPipeline
}

var ErrValidation = errors.New("validation failed")

func NewSiteActivationService(pipeline SiteActivationPipeline) SiteActivationService {
	return SiteActivationService{pipeline: pipeline}
}

func (s SiteActivationService) Activate(site domain.Site) error {
	if err := validateSiteActivation(site); err != nil {
		return err
	}
	return s.pipeline.Run(site)
}

func validateSiteActivation(site domain.Site) error {
	if strings.TrimSpace(site.ID) == "" {
		return errors.Join(ErrValidation, errors.New("site id is required"))
	}
	if strings.TrimSpace(site.Domain) == "" {
		return errors.Join(ErrValidation, errors.New("site domain is required"))
	}
	if site.OwnerUID <= 0 {
		return errors.Join(ErrValidation, errors.New("site owner uid must be positive"))
	}
	return nil
}
