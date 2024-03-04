package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type Service interface {
	FindCommits(t *testing.T, branch string, sinceSHA string) []*CommitFound
	WriteFile(t *testing.T, path, content, message string)
	WriteFileE(path, content, message string) error
}

type RepoService struct {
	Repos *GitHubRepositories
	Owner string
	Repo  string
}

func (s *RepoService) FindCommits(t *testing.T, branch string, sinceSHA string) []*CommitFound {
	commits, err := s.Repos.FindCommits(context.Background(), Base{
		Owner:  s.Owner,
		Repo:   s.Repo,
		Branch: branch,
	}, sinceSHA)
	if err != nil {
		t.Fatalf("unable to find commits: %v", err)
	}

	return commits
}

func (s *RepoService) WriteFile(t *testing.T, path, content, message string) {
	t.Helper()
	require.NoError(t, s.WriteFileE(path, content, message), "unable to write file")
}

func (s *RepoService) WriteFileE(path, content, message string) error {
	return s.Repos.Push(
		context.Background(),
		Base{
			Owner:  s.Owner,
			Repo:   s.Repo,
			Branch: "main",
		},
		Head{
			Branch: "main",
		},
		ChangeSet{
			Files: []*File{
				{
					Path:          path,
					ContentString: content,
				},
			},
			Message:   message,
			UserName:  "testkit",
			UserEmail: "",
		},
	)
}

var _ Service = &RepoService{}
