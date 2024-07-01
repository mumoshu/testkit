package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var log Logger = &namedLeveled{
		name: "test",
		l:    &Leveled{},
	}

	stdout, err := os.CreateTemp("", "stdout")
	require.NoError(t, err)

	stderr, err := os.CreateTemp("", "stderr")
	require.NoError(t, err)

	origStdout := os.Stdout
	origStderr := os.Stderr
	os.Stdout = stdout
	os.Stderr = stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	log.Debugf("foo")
	log.Infof("bar")
	log.Errorf("baz")

	require.NoError(t, stdout.Close())
	require.NoError(t, stderr.Close())

	so, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(so))

	se, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Equal(t, "[test info] bar\n[test error] baz\n", string(se))
}
