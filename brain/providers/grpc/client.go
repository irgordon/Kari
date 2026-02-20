package grpc

import (
	"errors"

	"github.com/kari/brain/internal/domain"
)

var errActionRejected = errors.New("agent rejected action")

type Client struct {
	transport Transport
}

func NewClient(address string) Client {
	return Client{transport: NewUnimplementedTransport(address)}
}

func NewClientWithTransport(transport Transport) Client {
	return Client{transport: transport}
}

func (c Client) CreateSystemUser(site domain.Site) error {
	return c.activateSite(site, ActionCreateSystemUser)
}

func (c Client) ApplyHTTPVHost(site domain.Site) error {
	return c.activateSite(site, ActionApplyHTTPVHost)
}

func (c Client) IssueCertificate(site domain.Site) error {
	return c.activateSite(site, ActionIssueCertificate)
}

func (c Client) ApplyHTTPSVHost(site domain.Site) error {
	return c.activateSite(site, ActionApplyHTTPSVHost)
}

func (c Client) RunSystemCheck(server domain.Server) (domain.SystemCheckReport, error) {
	response, err := c.transport.RunSystemCheck(RunSystemCheckRequest{ServerID: server.ID})
	if err != nil {
		return domain.SystemCheckReport{}, err
	}
	return domain.SystemCheckReport{
		Distro:         response.Distro,
		Version:        response.Version,
		Services:       response.Services,
		FirewallType:   response.FirewallType,
		FirewallStatus: response.FirewallStatus,
	}, nil
}

func (c Client) activateSite(site domain.Site, action SiteActivationAction) error {
	response, err := c.transport.ActivateSite(mapActivateSiteRequest(site, action))
	if err != nil {
		return err
	}
	if !response.OK {
		return errActionRejected
	}
	return nil
}

func mapActivateSiteRequest(site domain.Site, action SiteActivationAction) ActivateSiteRequest {
	return ActivateSiteRequest{
		SiteID:   site.ID,
		Domain:   site.Domain,
		IPv4:     site.IPv4,
		IPv6:     site.IPv6,
		OwnerUID: site.OwnerUID,
		Action:   action,
	}
}
