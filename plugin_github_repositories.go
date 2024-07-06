package testkit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mumoshu/testkit/git"
)

// GitHubWritableRepositoriesEnvProvider is a provider that behaves as a registory of writeable GitHub repositories.
// The repositories are reigstered via TESTKIT_GITHUB_WRITEABLE_REPOS variable, which is a comma-separated list of `owner/name` of GitHub repositories.
// The provider implementation never creates a new repository.
//
// The repository to be returned needs to be writeable,
// which means that the repository's testkit-config branch contains ".testkit.writeable" file in its root directory.
//
// The branch and the file requirement is there to avoid accidentally modifying the repository that is not intended to be modified.
//
// This is useful for testing the behavior of the code that modifies the repository AND
// you don't want to accidentally modify the repository that is not intended to be modified.
type GitHubWritableRepositoriesEnvProvider struct {
}

var _ Provider = &GitHubWritableRepositoriesEnvProvider{}
var _ GitHubWritableRepositoriesProvider = &GitHubWritableRepositoriesEnvProvider{}

func (p *GitHubWritableRepositoriesEnvProvider) Setup() error {
	var found bool
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "TESTKIT_") {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no TESTKIT_* environment variables found")
	}

	return nil
}

func (p *GitHubWritableRepositoriesEnvProvider) Cleanup() error {
	return nil
}

func (p *GitHubWritableRepositoriesEnvProvider) GetGitHubWritableRepository(opts ...GitHubWritableRepositoryOption) (*GitHubWritableRepository, error) {
	var conf GitHubWritableRepositoryConfig

	for _, opt := range opts {
		opt(&conf)
	}

	const (
		EnvRepo  = "TESTKIT_GITHUB_WRITEABLE_REPOS"
		EnvToken = "TESTKIT_GITHUB_TOKEN"
	)

	repos := os.Getenv(EnvRepo)
	if repos == "" {
		return nil, fmt.Errorf("%s environment variable is not set", EnvRepo)
	}

	token := os.Getenv(EnvToken)
	if token == "" {
		return nil, fmt.Errorf("%s environment variable is not set", EnvToken)
	}

	svc := &git.GitHubRepositories{
		Token: token,
	}

	for _, r := range strings.Split(repos, ",") {
		split := strings.Split(r, "/")
		owner, repo := split[0], split[1]
		b := git.Base{
			Owner:  owner,
			Repo:   repo,
			Branch: "testkit-config",
		}
		contentBytes, err := svc.GetFileContent(context.Background(), b, ".testkit.writable")
		if err != nil {
			return nil, fmt.Errorf("failed to get .testkit.writable file content from %s. This could be due to invalid or expired git credentials or network problem: %v", r, err)
		}

		contentStr := strings.TrimSpace(string(contentBytes))

		writeable := contentStr == "true"

		if !writeable {
			return nil, fmt.Errorf("repository %s has .testkit.writable file, but its content is not 'true' but %q", r, contentStr)
		}

		repoSvc := &git.RepoService{
			Repos: svc,
			Owner: owner,
			Repo:  repo,
		}

		return &GitHubWritableRepository{
			Name:    r,
			Token:   token,
			Service: repoSvc,
		}, nil
	}

	return nil, fmt.Errorf(
		"no writable repository found out of %q: At least one of the repositories must have .testkit.writable file",
		repos,
	)
}
