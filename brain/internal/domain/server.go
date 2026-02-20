package domain

type Server struct {
	ID      string
	Address string
}

type SystemCheckReport struct {
	Distro         string
	Version        string
	Services       map[string]string
	FirewallType   string
	FirewallStatus string
}
