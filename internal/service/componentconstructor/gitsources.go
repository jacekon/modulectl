package componentconstructor

import (
	"fmt"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types/component"
)

type GitService interface {
	GetLatestCommit(gitRepoPath string) (string, error)
}

type GitSourcesService struct {
	gitService GitService
}

func NewGitSourcesService(gitService GitService) (*GitSourcesService, error) {
	if gitService == nil {
		return nil, fmt.Errorf("gitService must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &GitSourcesService{
		gitService: gitService,
	}, nil
}

func (s *GitSourcesService) AddGitSourcesToConstructor(constructor *component.Constructor,
	gitRepoPath, gitRepoURL string,
) error {
	latestCommit, err := s.gitService.GetLatestCommit(gitRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %w", err)
	}

	constructor.AddGitSource(gitRepoURL, latestCommit)
	return nil
}
