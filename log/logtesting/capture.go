package logtesting

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type C struct {
	stdout *os.File
	stderr *os.File
}

func Capture(t *testing.T) *C {
	t.Helper()

	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	require.NoError(t, err)

	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	require.NoError(t, err)

	origStdout := os.Stdout
	origStderr := os.Stderr
	os.Stdout = stdout
	os.Stderr = stderr
	t.Cleanup(func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	})

	return &C{
		stdout: stdout,
		stderr: stderr,
	}
}

func (c *C) Stdout(t *testing.T) string {
	t.Helper()

	return c.str(t, c.stdout)
}

func (c *C) Stderr(t *testing.T) string {
	t.Helper()

	return c.str(t, c.stderr)
}

func (c *C) str(t *testing.T, f *os.File) string {
	t.Helper()

	require.NoError(t, f.Close(), "unable to close file %s", f.Name())

	data, err := os.ReadFile(f.Name())
	require.NoError(t, err, "unable to read file %s", f.Name())

	return string(data)
}
