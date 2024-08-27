package logtesting

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var stdout, stderr *os.File

func TestMain(m *testing.M) {
	// setup
	stdout = os.Stdout
	stderr = os.Stderr
	// run tests
	code := m.Run()
	// teardown
	os.Exit(code)
}

func TestCapture(t *testing.T) {
	t.Run("captures stdout and stderr", func(t *testing.T) {
		require.Equal(t, stdout, os.Stdout)
		require.Equal(t, stderr, os.Stderr)

		var c = Capture(t)

		require.NotEqual(t, stdout, os.Stdout)
		require.NotEqual(t, stderr, os.Stderr)

		fmt.Fprintln(os.Stdout, "foo")
		fmt.Fprintln(os.Stderr, "bar")
		fmt.Fprintln(os.Stderr, "baz")

		require.Equal(t, "foo\n", c.Stdout(t))
		require.Equal(t, "bar\nbaz\n", c.Stderr(t))

		require.NotEqual(t, stdout, os.Stdout)
		require.NotEqual(t, stderr, os.Stderr)
	})

	t.Run("cleans up the captured stdout and stderr", func(t *testing.T) {
		require.Equal(t, stdout, os.Stdout)
		require.Equal(t, stderr, os.Stderr)
	})
}

func ExampleCapture() {
	var stdout, stderr *os.File = os.Stdout, os.Stderr

	var (
		t = &testing.T{}
		c = Capture(t)
	)

	fmt.Fprintln(os.Stdout, "foo")
	fmt.Fprintln(os.Stderr, "bar")
	fmt.Fprintln(os.Stderr, "baz")

	fmt.Fprintf(stdout, "stdout swapped: %t\n", stdout != os.Stdout)
	fmt.Fprintf(stdout, "stderr swapped: %t\n", stderr != os.Stderr)
	fmt.Fprintf(stdout, "stdout: "+c.Stdout(t))
	fmt.Fprintf(stdout, "stderr: "+c.Stderr(t))

	// Output:
	// stdout swapped: true
	// stderr swapped: true
	// stdout: foo
	// stderr: bar
	// baz
}
