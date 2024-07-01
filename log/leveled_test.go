package log_test

import (
	"os"
	"testing"

	"github.com/mumoshu/testkit/log"
	"github.com/stretchr/testify/require"
)

func TestLeveled_default(t *testing.T) {
	d := t.TempDir()
	stdout, err := os.CreateTemp(d, "stdout")
	require.NoError(t, err)
	stderr, err := os.CreateTemp(d, "stderr")
	require.NoError(t, err)

	origStdout := os.Stdout
	origStderr := os.Stderr
	os.Stdout = stdout
	os.Stderr = stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	l := log.Leveled{}
	l.Logf("a", log.Info, "foo")
	l.Logf("b", log.Debug, "bar")
	l.Logf("c", log.Error, "baz")

	require.NoError(t, stdout.Close())
	require.NoError(t, stderr.Close())

	so, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(so))

	se, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Equal(t, "[a info] foo\n[c error] baz\n", string(se))
}
