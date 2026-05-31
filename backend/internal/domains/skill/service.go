package skill

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	"dotblue/internal/domains/agent"
	"github.com/google/uuid"
)

var skillCodePattern = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)

// Service remains the single facade for the skill bounded context.
// Internally, its methods are split across catalog, lifecycle, availability,
// binding, and hub-focused files so platform and enterprise concerns stay
// unified in model while becoming easier to evolve.
type Service struct {
	repo        Repository
	idGenerator func() string
	now         func() time.Time
	loadAgent   func(id string) (*agent.Agent, error)
	fetchURL    func(url string) ([]byte, error)
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:        repo,
		idGenerator: uuid.NewString,
		now:         time.Now,
		loadAgent:   agent.GetById,
		fetchURL: func(url string) ([]byte, error) {
			resp, err := http.Get(url) // #nosec G107 - skill hub URLs are platform-admin managed.
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, errors.New(resp.Status)
			}
			return io.ReadAll(resp.Body)
		},
	}
}
