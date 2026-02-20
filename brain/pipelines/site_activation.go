package pipelines

import (
	"github.com/kari/brain/internal/domain"
	"github.com/kari/brain/internal/ports"
)

type SiteActivationPipeline struct {
	agent      ports.Agent
	dns        ports.DNSProvider
	repository ports.SiteRepository
}

func NewSiteActivationPipeline(agent ports.Agent, dns ports.DNSProvider, repository ports.SiteRepository) SiteActivationPipeline {
	return SiteActivationPipeline{agent: agent, dns: dns, repository: repository}
}

func (p SiteActivationPipeline) Run(site domain.Site) error {
	if err := p.agent.CreateSystemUser(site); err != nil {
		return err
	}
	if err := p.agent.ApplyHTTPVHost(site); err != nil {
		return err
	}
	if err := p.dns.EnsureAddressRecords(site); err != nil {
		return err
	}
	if err := p.agent.IssueCertificate(site); err != nil {
		return err
	}
	if err := p.agent.ApplyHTTPSVHost(site); err != nil {
		return err
	}
	return p.repository.MarkActive(site.ID)
}
