package ports

import "github.com/kari/brain/internal/domain"

type DNSProvider interface {
	EnsureAddressRecords(site domain.Site) error
}
