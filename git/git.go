package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// File represents the desired state of a file in a repository.
// The service should create a commit that contains the file described by this struct.
//
// The service should not create a commit if the file already exists and the content of the file
// is the same as the content described by this struct.
//
// The service should create a commit if the file already exists and the content of the file
// is different from the content described by this struct.
//
// The Path field is required.
// There must be only one of ContentString, ContentBytes, and ContentFromLocalFile set.
type File struct {
	// Path is a path to the file within the repository.
	Path string
	// ContentString is a string that contains the content of the file.
	ContentString string
	// ContentBytes is a byte slice that contains the content of the file.
	ContentBytes []byte
	// ContentFromLocalFile is a path to a local file that contains the content of the file.
	ContentFromLocalFile string

	// ContentFunc is a function that returns the content of the file.
	// The argument is the content of the file that already exists in the repository.
	// The argument is nil if the file does not exist.
	// The return value is the content of the file that should be committed.
	ContentBytesFunc func([]byte) ([]byte, error)
}

// Head represents a head branch to be created and optionally a tag to be created.
type Head struct {
	Branch string
	Tag    string
}

// ChangeSet represents a commit to be created and pushed.
type ChangeSet struct {
	// Files is a list of files to be created and pushed.
	Files []*File

	// Message is a commit message.
	Message string

	// UserName is user.name of the committer.
	UserName string
	// UserEmail is user.email of the committer.
	UserEmail string
}

type CloneOptions struct {
	// Token is a GitHub API token used to clone the repository.
	Token string
}

// Base represents a base repository and branch.
type Base struct {
	Owner  string
	Repo   string
	Branch string
}

func (b Base) CloneGitURL() string {
	return fmt.Sprintf("git@github.com:%s/%s.git", b.Owner, b.Repo)
}

func (b Base) CloneHTTPSURL(token string) string {
	return fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, b.Owner, b.Repo)
}

func CloneIntoNewBranch(base Base, local, branch string, opts CloneOptions) error {
	var repoURL string
	if opts.Token != "" {
		repoURL = base.CloneHTTPSURL(opts.Token)
	} else {
		repoURL = base.CloneGitURL()
	}

	if _, err := git("", "clone", repoURL, local); err != nil {
		return fmt.Errorf("git-clone: %s", err)
	}

	if _, err := git(local, "checkout", "-b", branch, "origin/"+base.Branch); err != nil {
		return fmt.Errorf("git-checkout: %w", err)
	}

	return nil
}

func WriteAndAddFiles(local string, files []*File) error {
	for _, f := range files {
		if err := WriteAndAddFile(local, f); err != nil {
			return err
		}
	}

	return nil
}

func CommitRenameBranchAndPush(local string, commit ChangeSet, head Head) error {
	if commit.UserName != "" {
		if _, err := git(local, "config", "user.name", commit.UserName); err != nil {
			return err
		}
	}

	if commit.UserEmail != "" {
		if _, err := git(local, "config", "user.email", commit.UserEmail); err != nil {
			return err
		}
	}

	message := commit.Message
	if message == "" {
		message = "Automated commit via testkit"
	}

	if _, err := git(local, "commit", "-m", message); err != nil {
		return err
	}

	branch := head.Branch
	if branch == "" {
		return fmt.Errorf("head branch must be set")
	}

	if _, err := git(local, "branch", "-m", branch); err != nil {
		return err
	}

	if _, err := git(local, "push", "origin", branch); err != nil {
		return err
	}

	if head.Tag != "" {
		if _, err := git(local, "tag", head.Tag); err != nil {
			return err
		}

		if _, err := git(local, "push", "origin", head.Tag); err != nil {
			return err
		}
	}

	return nil
}

func WriteAndAddFile(local string, f *File) error {
	var content []byte
	if f.ContentString != "" {
		content = []byte(f.ContentString)
	} else if f.ContentBytes != nil {
		content = f.ContentBytes
	} else if f.ContentFromLocalFile != "" {
		var err error
		content, err = os.ReadFile(f.ContentFromLocalFile)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("one of ContentString, ContentBytes, and ContentFromLocalFile must be set")
	}

	if err := os.MkdirAll(filepath.Dir(filepath.Join(local, f.Path)), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(local, f.Path), content, 0644); err != nil {
		return err
	}

	if _, err := git(local, "add", f.Path); err != nil {
		return err
	}

	return nil
}

// CommitFound contains the information about a commit that was found in the repository.
//
// This contains both the subset of information obtained from GitHub API
// and some derived information about the files that were affected by the commit.
// The affected files, named ChangeSet, contains paths and contents of the affected files,
// so that the caller can use this information to validate the state of the files in the repository,
// after some operations are done.
type CommitFound struct {
	SHA       string
	Message   string
	ChangeSet ChangeSet
}

func FindCommits(local string, sinceSHA string) ([]*CommitFound, error) {
	var r string
	if sinceSHA == "" {
		r = "HEAD"
	} else {
		r = sinceSHA + "..HEAD"
	}
	// This returns a new-line-delimited list of commits in the format "SHA1 commit message"
	// from the newest to the oldest.
	commits, err := git(local, "log", "--pretty=format:%H %s", r)
	if err != nil {
		return nil, err
	}

	return convertCommitSHAsToStateOfAffectedFiles(commits)
}

func convertCommitSHAsToStateOfAffectedFiles(commits string) ([]*CommitFound, error) {
	lines := strings.Split(commits, "\n")
	commitFounds := make([]*CommitFound, 0, len(lines))
	for _, line := range lines {
		a := strings.SplitN(line, " ", 2)
		sha, message := a[0], a[1]
		commit := &CommitFound{
			SHA: sha,
			ChangeSet: ChangeSet{
				Message: message,
			},
		}

		authorName, authorEmail, files, err := getCommitDetails(sha)
		if err != nil {
			return nil, err
		}

		commit.ChangeSet.UserName = authorName
		commit.ChangeSet.UserEmail = authorEmail
		commit.ChangeSet.Files = files

		commitFounds = append(commitFounds, commit)
	}

	return commitFounds, nil
}

func getCommitDetails(sha string) (string, string, []*File, error) {
	authorName, err := git("", "show", "-s", "--format=%an", sha)
	if err != nil {
		return "", "", nil, err
	}

	authorEmail, err := git("", "show", "-s", "--format=%ae", sha)
	if err != nil {
		return "", "", nil, err
	}

	// This gives us a list of files in the format "status path"
	// where status is "A" for added, "M" for modified, and "D" for deleted.
	// The list is separated by new lines.
	// Example:
	//   M       kubectl.go
	nameStatuses, err := git("", "show", "--pretty=format:", "--name-status", sha)
	if err != nil {
		return "", "", nil, err
	}

	files := make([]*File, 0)
	for _, nameStatus := range strings.Split(nameStatuses, "\n") {
		if nameStatus == "" {
			continue
		}

		a := strings.SplitN(nameStatus, "\t", 2)
		status, path := a[0], a[1]

		var content string

		switch status {
		case "A":
			content, err = showFileContentAtCommit(sha, path)
			if err != nil {
				return "", "", nil, err
			}
		case "M":
			content, err = showFileContentAtCommit(sha, path)
			if err != nil {
				return "", "", nil, err
			}
		case "D":
			// We don't need to do anything for deleted files.
			continue
		default:
			return "", "", nil, fmt.Errorf("unknown status: %s", status)
		}

		files = append(files, &File{
			Path:          path,
			ContentString: content,
		})
	}

	return authorName, authorEmail, files, nil
}

func showFileContentAtCommit(sha, path string) (string, error) {
	content, err := git("", "show", fmt.Sprintf("%s:%s", sha, path))
	if err != nil {
		return "", err
	}

	return content, nil
}

// git runs the git command with the given arguments and returns the output containing
// the stdout and stderr.
//
// An error is returned if the command fails.
// The error message should contain the stdout and stderr of the command to give the caller
// information about what went wrong.
func git(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		errWithOutput := fmt.Errorf("git command failed with error: %w", err)
		errWithOutput = fmt.Errorf("%s\n\n%s", errWithOutput, string(out))
		return string(out), errWithOutput
	}
	return string(out), nil
}
