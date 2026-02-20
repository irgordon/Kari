package inmemory

import "github.com/kari/brain/internal/domain"

type SystemChecker struct{}

func NewSystemChecker() SystemChecker {
	return SystemChecker{}
}

func (SystemChecker) RunSystemCheck(_ domain.Server) (domain.SystemCheckReport, error) {
	return domain.SystemCheckReport{
		Distro:  "ubuntu",
		Version: "22.04",
		Services: map[string]string{
			"nginx":   "running",
			"php-fpm": "running",
		},
		FirewallType:   "ufw",
		FirewallStatus: "active",
	}, nil
}
