package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
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

	var p Printer = &defaultPrinter{}
	p.Printf("hello, %s", "world")

	require.NoError(t, stdout.Close())
	require.NoError(t, stderr.Close())

	so, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(so))

	se, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Equal(t, "hello, world\n", string(se))
}
