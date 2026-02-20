package ports

type SiteRepository interface {
	MarkActive(siteID string) error
}
