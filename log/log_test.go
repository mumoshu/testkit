package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	os.Setenv(Env, "baz=debug")
	defer os.Unsetenv(Env)

	log := New()

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

	log.For("foo").Infof("Hello, %s", "world")

	bar := log.For("bar")
	bar.Infof("Did %s", "a")
	bar.Errorf("Failed to %s", "a")

	// log.For returns a Logger which can be a struct field.
	// This is useful for writing logging code in methods a less verbose way.
	baz := struct{ Logger }{log.For("baz")}
	baz.Debugf("About to do %s", "b")
	baz.Infof("Did %s", "b")
	baz.Errorf("Failed to %s", "b")

	require.NoError(t, stdout.Close())
	require.NoError(t, stderr.Close())

	so, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(so))

	se, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Equal(t, "[foo info] Hello, world\n[bar info] Did a\n[bar error] Failed to a\n[baz debug] About to do b\n[baz info] Did b\n[baz error] Failed to b\n", string(se))
}

func TestL(t *testing.T) {
	os.Setenv(Env, "baz=debug")
	defer os.Unsetenv(Env)

	app := struct{ L }{}

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

	app.Infof("Did %s", "a")

	app.Logger = Default.For("app")
	app.Infof("Did %s", "b")

	require.NoError(t, stdout.Close())
	require.NoError(t, stderr.Close())

	so, err := os.ReadFile(stdout.Name())
	require.NoError(t, err)
	require.Equal(t, "", string(so))

	se, err := os.ReadFile(stderr.Name())
	require.NoError(t, err)
	require.Equal(t, "[github.com/mumoshu/testkit/log.TestL info] Did a\n[app info] Did b\n", string(se))
}
