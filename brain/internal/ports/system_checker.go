package ports

import "github.com/kari/brain/internal/domain"

type SystemChecker interface {
	RunSystemCheck(server domain.Server) (domain.SystemCheckReport, error)
}
