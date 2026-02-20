package grpc

import "errors"

type Transport interface {
	RunSystemCheck(request RunSystemCheckRequest) (RunSystemCheckResponse, error)
	ActivateSite(request ActivateSiteRequest) (ActivateSiteResponse, error)
}

type RunSystemCheckRequest struct {
	ServerID string
}

type RunSystemCheckResponse struct {
	Distro         string
	Version        string
	Services       map[string]string
	FirewallType   string
	FirewallStatus string
}

type ActivateSiteRequest struct {
	SiteID   string
	Domain   string
	IPv4     string
	IPv6     string
	OwnerUID int
	Action   SiteActivationAction
}

type ActivateSiteResponse struct {
	OK bool
}

type SiteActivationAction string

const (
	ActionCreateSystemUser SiteActivationAction = "CREATE_SYSTEM_USER"
	ActionApplyHTTPVHost   SiteActivationAction = "APPLY_HTTP_VHOST"
	ActionIssueCertificate SiteActivationAction = "ISSUE_CERTIFICATE"
	ActionApplyHTTPSVHost  SiteActivationAction = "APPLY_HTTPS_VHOST"
)

type unimplementedTransport struct {
	address string
}

func NewUnimplementedTransport(address string) Transport {
	return unimplementedTransport{address: address}
}

func (t unimplementedTransport) RunSystemCheck(_ RunSystemCheckRequest) (RunSystemCheckResponse, error) {
	return RunSystemCheckResponse{}, t.notImplemented("RunSystemCheck")
}

func (t unimplementedTransport) ActivateSite(_ ActivateSiteRequest) (ActivateSiteResponse, error) {
	return ActivateSiteResponse{}, t.notImplemented("ActivateSite")
}

func (t unimplementedTransport) notImplemented(method string) error {
	return errors.New("grpc transport method not implemented: " + method + " (address: " + t.address + ")")
}
