package testkit

import (
	"testing"

	"github.com/mumoshu/testkit/git"
)

type GitHubWritableRepositoriesProvider interface {
	GetGitHubWritableRepository(opts ...GitHubWritableRepositoryOption) (*GitHubWritableRepository, error)
}

type GitHubWritableRepository struct {
	ID string
	// Name is the name of the repository.
	// This is the name that appears in the URL.
	// For example, if the URL is
	// https://github.com/mumoshu/example-repo,
	// then the name is "mumoshu/example-repo".
	Name string

	// Token is the GitHub personal access token.
	// This is used to authenticate to GitHub API for interacting
	// with the repository.
	Token string

	git.Service
}

type GitHubWritableRepositoryConfig struct {
	ID string
}

type GitHubWritableRepositoryOption func(*GitHubWritableRepositoryConfig)

// GitHubWritableRepository retrieves a writable GitHub repository.
//
// It iterates over the available providers and calls the GetGitHubWritableRepository method on each provider.
// If no provider implements GetGitHubWritableRepository, it fails the test.
// If multiple providers implement GetGitHubWritableRepository, it returns the first successful one.
// If multiple providers implement GetGitHubWritableRepository and all of them fail, it fails the test.
func (tk *TestKit) GitHubWritableRepository(t *testing.T, opts ...GitHubWritableRepositoryOption) *GitHubWritableRepository {
	t.Helper()

	var cp GitHubWritableRepositoriesProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(GitHubWritableRepositoriesProvider)
		if ok {
			ghRepo, err := cp.GetGitHubWritableRepository(opts...)
			if err != nil {
				t.Logf("unable to get GitHub writable repository: %v", err)
				continue
			}

			return ghRepo
		}
	}

	if cp == nil {
		t.Fatal("no GitHubWritableRepositoriesProvider found")
	}

	return nil
}
