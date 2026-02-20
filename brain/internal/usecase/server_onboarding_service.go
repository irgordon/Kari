package usecase

import (
	"errors"
	"strings"

	"github.com/kari/brain/internal/domain"
	"github.com/kari/brain/internal/ports"
)

type ServerOnboardingService struct {
	systemChecker ports.SystemChecker
}

func NewServerOnboardingService(systemChecker ports.SystemChecker) ServerOnboardingService {
	return ServerOnboardingService{systemChecker: systemChecker}
}

func (s ServerOnboardingService) Onboard(server domain.Server) (domain.SystemCheckReport, error) {
	if err := validateServerOnboarding(server); err != nil {
		return domain.SystemCheckReport{}, err
	}
	return s.systemChecker.RunSystemCheck(server)
}

func validateServerOnboarding(server domain.Server) error {
	if strings.TrimSpace(server.ID) == "" {
		return errors.Join(ErrValidation, errors.New("server id is required"))
	}
	if strings.TrimSpace(server.Address) == "" {
		return errors.Join(ErrValidation, errors.New("server address is required"))
	}
	return nil
}
