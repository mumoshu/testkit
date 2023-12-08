package testkit

import "testing"

type GitHubRepositoryProvider interface {
	GetGitHubRepository(opts ...GitHubRepositoryOption) (*GitHubRepository, error)
}

type GitHubRepository struct {
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
}

type GitHubRepositoryConfig struct {
	ID string
}

type GitHubRepositoryOption func(*GitHubRepositoryConfig)

// GitHubRepository creates a GitHub repository.
// The repository is created by the provider implementation.
// The provider implementation may create a new repository,
// or may use an existing repository.
// The provider implementation may also create a new GitHub personal access token,
// or may use an existing one.
// It iterates over the available providers and calls the GetGitHubRepository method on each provider.
// If no provider implements GetGitHubRepository, it fails the test.
// If multiple providers implement GetGitHubRepository, it returns the first successful one.
// If multiple providers implement GetGitHubRepository and all of them fail, it fails the test.
func (tk *TestKit) GitHubRepository(t *testing.T, opts ...GitHubRepositoryOption) *GitHubRepository {
	t.Helper()

	var cp GitHubRepositoryProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(GitHubRepositoryProvider)
		if ok {
			ghRepo, err := cp.GetGitHubRepository(opts...)
			if err != nil {
				t.Logf("unable to get GitHub repository: %v", err)
				continue
			}

			return ghRepo
		}
	}

	if cp == nil {
		t.Fatal("no GitHubRepositoryProvider found")
	}

	return nil
}
