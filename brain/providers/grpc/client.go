package grpc

import (
	"errors"

	"github.com/kari/brain/internal/domain"
)

type Client struct {
	address string
}

func NewClient(address string) Client {
	return Client{address: address}
}

func (c Client) CreateSystemUser(_ domain.Site) error {
	return c.notImplemented("CreateSystemUser")
}

func (c Client) ApplyHTTPVHost(_ domain.Site) error {
	return c.notImplemented("ApplyHTTPVHost")
}

func (c Client) IssueCertificate(_ domain.Site) error {
	return c.notImplemented("IssueCertificate")
}

func (c Client) ApplyHTTPSVHost(_ domain.Site) error {
	return c.notImplemented("ApplyHTTPSVHost")
}

func (c Client) RunSystemCheck(_ domain.Server) (domain.SystemCheckReport, error) {
	return domain.SystemCheckReport{}, c.notImplemented("RunSystemCheck")
}

func (c Client) notImplemented(method string) error {
	return errors.New("grpc client method not implemented: " + method + " (address: " + c.address + ")")
}
