package inmemory

import "github.com/kari/brain/internal/domain"

type Agent struct{}

func NewAgent() Agent {
	return Agent{}
}

func (Agent) CreateSystemUser(_ domain.Site) error {
	return nil
}

func (Agent) ApplyHTTPVHost(_ domain.Site) error {
	return nil
}

func (Agent) IssueCertificate(_ domain.Site) error {
	return nil
}

func (Agent) ApplyHTTPSVHost(_ domain.Site) error {
	return nil
}
