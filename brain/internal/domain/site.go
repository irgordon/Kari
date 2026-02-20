package domain

type Site struct {
	ID       string
	Domain   string
	IPv4     string
	IPv6     string
	OwnerUID int
}
