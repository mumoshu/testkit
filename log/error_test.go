package log_test

import (
	"testing"

	"github.com/mumoshu/testkit/error"
	"github.com/mumoshu/testkit/log"
	"github.com/mumoshu/testkit/log/logtesting"
	"github.com/stretchr/testify/require"
)

func TestLogError(t *testing.T) {
	var e = error.New(
		"unable to get file from git repository `foo`",
		error.Source("some.json", 2, "some code"),
		error.Long("This application requires access to the file in the repository `foo` for reading the configuration, but was unable to do so due to an internal error."),
		error.Remediation(`Please make sure that the repository is accessible to the application.
If you are using a private repository, please make sure that git is configured with the correct credentials.
If you are using git+ssh protocol, check if `+"`ssh $GITHUB_USER@github.com`"+` works.`,
		),
	)

	var logger = log.New()

	c := logtesting.Capture(t)

	log.E(logger, "test", log.Info, e)

	require.Empty(t, c.Stdout(t))
	require.Equal(t, `[test info] ╷
│ unable to get file from git repository `+"`foo`"+`
│
│   on some.json line 2:
│    2: some code
│
│ This application requires access to the file in the repository `+"`foo`"+` for reading the configuration, but was unable to do so due to an internal error.
│
│ Please make sure that the repository is accessible to the application.
If you are using a private repository, please make sure that git is configured with the correct credentials.
If you are using git+ssh protocol, check if `+"`ssh $GITHUB_USER@github.com`"+` works.
╵
`, c.Stderr(t))
}
