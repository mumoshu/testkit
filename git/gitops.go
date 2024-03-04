package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// GitHubRepositories provides a set of operations to interact with repositories.
// In an abstract sense, this is used to trigger a side effect in a repository, or to validate the outcome of a side effect.
//
// The supported operations are:
// - Create and push a commit to the repository
// - Create and push a tag to the repository
// - Create and push a branch to the repository
// - Create and send a pull request to the repository
// - Create a release in the repository
// - Send a comment to a pull request in the repository
// - Send a repository dispatch event to the repository
// - Merge a pull request in the repository
// - Find a commit in the repository
// - Find a tag in the repository
// - Find a pull request in the repository
// - Find a file in the repository (This is used by the provider to validate the repository)
//
// The unsupported operations are:
// - Validate the repository (e.g. if it's a repository supposed to be reset and used by a test)
// - Reset the repository (e.g. delete all the branches and tags)
// - Find a GitHub Actions workflow run in the repository (You check the outcome of the workflow run instead)
//
// These unsupported operations are implemented in the provider, not this abstraction,
// because they are provider-specific.
// The consumer of this library should never directly do these operations because any mistake
// could cause catastrophic damage to an unintended repository.
//
// This service uses the GitHub API and the git command under the hood.
// Unless otherwise noted, it's assumed that the GitHub API and the git command are available.
// If the GitHub API or the git command is not available, the service should return an error on
// any operation.
//
// For calling various GitHub API and interacting with a private or an internal repository,
// the service should use the GitHub API token.
type GitHubRepositories struct {
	// Token is a GitHub API token that is used to call various GitHub API
	// and interact with a private or an internal repository.
	// If empty, the service reads the token from the GITHUB_TOKEN environment variable.
	Token string

	// TempDir is a path to a temporary directory that is used to clone the repository.
	// The service creates a subdirectory under this directory and clones the repository into it.
	//
	// If empty, the service uses the system's temporary directory.
	TempDir string

	// RetainClonedRepository specifies whether the service should retain the cloned repository
	// after the service is done.
	// This is usually set to true when the consumer of the library wants to inspect the cloned
	// repository after the service is done, perhaps for debugging.
	RetainClonedRepository bool

	// clonedRepoIndex is an index that is used to generate a unique directory name for a cloned
	// repository.
	clonedRepoIndex int

	// workBranchIndex is an index that is used to generate a unique branch name for a work branch.
	workBranchIndex int
}

// PullRequest contains the information about a pull request to be created and sent to a repository.
type PullRequest struct {
	Title  string
	Body   string
	Labels []string
}

// Send pushes the given commit to the repository and creates a pull request.
func (s *GitHubRepositories) Send(ctx context.Context, base Base, head Head, commit ChangeSet, pr PullRequest) (*PullRequestCreated, error) {
	clean, err := s.push(base, head, commit)
	if err != nil {
		return nil, err
	}

	if clean != nil {
		defer clean()
	}

	r, err := s.createPullRequest(ctx, base, head, pr)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Comment sends a comment to the pull request in the repository.
func (s *GitHubRepositories) Comment(ctx context.Context, base Base, prNumber int, body string) error {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	_, _, err := client.Issues.CreateComment(ctx, base.Owner, base.Repo, prNumber, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("sending comment: %w", err)
	}

	return nil
}

// Push the changeset to a repository.
//
// This does so by cloning the base repository, writes a commit representing the change set, and pushes it to the branch, optionally tagging it.
func (s *GitHubRepositories) Push(ctx context.Context, base Base, head Head, commit ChangeSet) error {
	clean, err := s.push(base, head, commit)

	if clean != nil {
		defer clean()
	}

	return err
}

func (s *GitHubRepositories) Dispatch(ctx context.Context, base Base, eventType string, clientPayload interface{}) error {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	payload, err := json.Marshal(clientPayload)
	if err != nil {
		return err
	}

	raw := json.RawMessage(payload)

	_, _, err = client.Repositories.Dispatch(ctx, base.Owner, base.Repo, github.DispatchRequestOptions{
		EventType:     eventType,
		ClientPayload: &raw,
	})
	if err != nil {
		return fmt.Errorf("dispatching event: %w", err)
	}

	return nil
}

// Merge merges the pull request in the repository.
func (s *GitHubRepositories) Merge(ctx context.Context, base Base, prNumber int) error {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	_, _, err := client.PullRequests.Merge(ctx, base.Owner, base.Repo, prNumber, "", &github.PullRequestOptions{
		MergeMethod: "merge",
	})
	if err != nil {
		return fmt.Errorf("merging pull request: %w", err)
	}

	return nil
}

// Release creates a release in the repository, along with a lightweight tag.
func (s *GitHubRepositories) Release(ctx context.Context, base Base, name string) error {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	_, _, err := client.Repositories.CreateRelease(ctx, base.Owner, base.Repo, &github.RepositoryRelease{
		TagName: &name,
	})
	if err != nil {
		return fmt.Errorf("creating release: %w", err)
	}

	return nil
}

// FindCommits finds a commit in the repository.
//
// This does so by cloning the base repository, and finding all the commits made since the given SHA.
// If the SHA is empty, it finds the latest commit in the given repository and branch.
//
// An expected use case is to call this method twice, with the first call to find the commit before the change set,
// and the second call to find the commit after the change set.
// This way, you can see if the change set is reflected in the repository.
func (s *GitHubRepositories) FindCommits(ctx context.Context, base Base, sinceSHA string) ([]*CommitFound, error) {
	local := s.newLocalRepoDirName(base)
	err := CloneIntoNewBranch(
		base,
		local,
		s.newWorkBranchName(),
		CloneOptions{
			Token: s.Token,
		},
	)
	if err != nil {
		return nil, err
	}

	commits, err := FindCommits(local, sinceSHA)
	if err != nil {
		return nil, err
	}

	if !s.RetainClonedRepository {
		_ = os.RemoveAll(local)
	}

	return commits, nil
}

// FindSemverTags finds semver tags in the repository.
//
// This does so by using GitHub API to find the tags in the repository.
// sinceVer is a semver version string, and the method finds the tags that are greater than or equal to the given version.
// If sinceVer is empty, it finds the latest tag in the given repository.
// The latest tag is the tag that is the greatest in the semver order.
//
// Either way, the caller can use the first tag in the returned list to find the latest tag in the repository,
// that is at least greater than or equal to the given sinceVer.
func (s *GitHubRepositories) FindSemverTags(ctx context.Context, base Base, sinceVer string) ([]string, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	var (
		// We paginate ListTags to find all the tags in the repository.
		tags []string

		since *semver.Version
	)

	if sinceVer != "" {
		var err error
		since, err = semver.NewVersion(sinceVer)
		if err != nil {
			return nil, fmt.Errorf("parsing semver: %w", err)
		}
	}

	opt := &github.ListOptions{PerPage: 100}
	for {
		ts, resp, err := client.Repositories.ListTags(ctx, base.Owner, base.Repo, opt)
		if err != nil {
			return nil, fmt.Errorf("listing tags: %w", err)
		}

		for _, t := range ts {
			tVer, err := semver.NewVersion(*t.Name)
			if err != nil {
				return nil, fmt.Errorf("parsing semver: %w", err)
			}

			if since == nil {
				if len(tags) == 0 {
					tags = append(tags, *t.Name)
				} else {
					oldVer, err := semver.NewVersion(tags[0])
					if err != nil {
						return nil, fmt.Errorf("parsing semver: %w", err)
					}

					if tVer.GreaterThan(oldVer) {
						tags[0] = *t.Name
					}
				}
				continue
			} else if tVer.GreaterThan(since) || tVer.Equal(since) {
				tags = append(tags, t.GetName())
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return tags, nil
}

// GetFileContent finds a file in the repository.
//
// This does so by cloning the base repository, and finding the file in the given branch.
func (s *GitHubRepositories) GetFileContent(ctx context.Context, base Base, path string) ([]byte, error) {
	local := s.newLocalRepoDirName(base)
	err := CloneIntoNewBranch(
		base,
		local,
		s.newWorkBranchName(),
		CloneOptions{
			Token: s.Token,
		},
	)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath.Join(local, path))
	if err != nil {
		return nil, err
	}

	if !s.RetainClonedRepository {
		_ = os.RemoveAll(local)
	}

	return content, nil
}

// PullRequestFound contains the information about a pull request found in the repository.
type PullRequestFound struct {
	Number  int
	BaseSHA string
	Commits []*CommitFound
}

// FindPullRequests finds pull requests in the repository.
//
// The last argument, sinceNumber is the pull request number, and the method finds the pull requests that are greater than or equal to the given number.
// If sinceNumber is 0, it finds the latest pull request in the given repository.
// The latest pull request is the pull request that is the greatest in the pull request number order.
//
// Either way, the caller can use the first pull request in the returned list to find the latest pull request in the repository,
// that is at least greater than or equal to the given sinceNumber.
func (s *GitHubRepositories) FindPullRequests(ctx context.Context, base Base, sinceNumber int) ([]*PullRequestFound, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	opt := &github.PullRequestListOptions{State: "all", ListOptions: github.ListOptions{PerPage: 100}}
	var (
		prs []*PullRequestFound
	)

	for {
		ps, resp, err := client.PullRequests.List(ctx, base.Owner, base.Repo, opt)
		if err != nil {
			return nil, fmt.Errorf("listing pull requests: %w", err)
		}

		for _, p := range ps {
			if sinceNumber == 0 {
				if len(prs) == 0 {
					prs = append(prs, &PullRequestFound{
						Number:  p.GetNumber(),
						BaseSHA: p.GetBase().GetSHA(),
					})
				} else {
					if p.GetNumber() > prs[0].Number {
						prs[0] = &PullRequestFound{
							Number:  p.GetNumber(),
							BaseSHA: p.GetBase().GetSHA(),
						}
					}
				}
				continue
			} else if p.GetNumber() >= sinceNumber {
				prs = append(prs, &PullRequestFound{
					Number:  p.GetNumber(),
					BaseSHA: p.GetBase().GetSHA(),
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	// Populate commits for each pull request
	for i := 0; i < len(prs); i++ {
		p := prs[i]

		ref := fmt.Sprintf("refs/pull/%d/head", p.Number)

		b := Base{
			Owner:  base.Owner,
			Repo:   base.Repo,
			Branch: ref,
		}

		commits, err := s.FindCommits(ctx, b, p.BaseSHA)
		if err != nil {
			return nil, err
		}

		if len(commits) == 0 {
			return nil, fmt.Errorf("no commits found in the pull request")
		}

		p.Commits = commits

		prs[i] = p
	}

	return prs, nil
}

// push clones the base repository, writes a commit representing the change set, and pushes it to the branch, optionally tagging it.
func (s *GitHubRepositories) push(base Base, head Head, commit ChangeSet) (func(), error) {
	local := s.newLocalRepoDirName(base)
	newBranch := s.newWorkBranchName()
	err := CloneIntoNewBranch(
		base,
		local,
		newBranch,
		CloneOptions{
			Token: s.Token,
		},
	)
	if err != nil {
		return nil, err
	}

	clean := func() {
		if !s.RetainClonedRepository {
			_ = os.RemoveAll(local)
		}
	}

	if err := WriteAndAddFiles(local, commit.Files); err != nil {
		return clean, err
	}

	if err := CommitRenameBranchAndPush(local, commit, head); err != nil {
		return clean, err
	}

	return clean, nil
}

func (s *GitHubRepositories) newLocalRepoDirName(base Base) string {
	s.clonedRepoIndex++

	return fmt.Sprintf("%s/%s/%s/%2d", s.tempDir(), base.Owner, base.Repo, s.clonedRepoIndex)
}

func (s *GitHubRepositories) newWorkBranchName() string {
	s.workBranchIndex++

	return fmt.Sprintf("testkit-work-%d", s.workBranchIndex)
}

func (s *GitHubRepositories) tempDir() string {
	if s.TempDir != "" {
		return s.TempDir
	}

	return filepath.Join(os.TempDir(), "testkit", "ghreposvc")
}

// PullRequestCreated contains the information about a pull request that was created and sent to a repository.
type PullRequestCreated struct {
	Number int
}

func (s *GitHubRepositories) createPullRequest(ctx context.Context, base Base, head Head, pr PullRequest) (*PullRequestCreated, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	))
	client := github.NewClient(httpClient)

	r, _, err := client.PullRequests.Create(ctx, base.Owner, base.Repo, &github.NewPullRequest{
		Title: &pr.Title,
		Body:  &pr.Body,
		Head:  &head.Branch,
		Base:  &base.Branch,
	})
	if err != nil {
		return nil, fmt.Errorf("creating pull request: %w", err)
	}

	if len(pr.Labels) > 0 {
		_, _, err = client.Issues.AddLabelsToIssue(ctx, base.Owner, base.Repo, r.GetNumber(), pr.Labels)
		if err != nil {
			return nil, fmt.Errorf("adding labels to pull request: %w", err)
		}
	}

	return &PullRequestCreated{Number: r.GetNumber()}, nil
}
