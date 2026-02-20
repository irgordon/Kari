package ports

import "github.com/kari/brain/internal/domain"

type Agent interface {
	CreateSystemUser(site domain.Site) error
	ApplyHTTPVHost(site domain.Site) error
	IssueCertificate(site domain.Site) error
	ApplyHTTPSVHost(site domain.Site) error
}
