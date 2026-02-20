package inmemory

import "github.com/kari/brain/internal/domain"

type DNSProvider struct{}

func NewDNSProvider() DNSProvider {
	return DNSProvider{}
}

func (DNSProvider) EnsureAddressRecords(_ domain.Site) error {
	return nil
}
