package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeveledCache(t *testing.T) {
	// The default log level is Info
	a := leveledCache{}
	require.True(t, a.Enabled("foo", Error))
	require.True(t, a.Enabled("foo", Info))
	require.False(t, a.Enabled("foo", Debug))
	require.True(t, a.Enabled("bar", Error))
	require.True(t, a.Enabled("bar", Info))
	require.False(t, a.Enabled("bar", Debug))

	// Log levels are read from the default environment variable
	oldEnv := os.Getenv(Env)
	os.Setenv(Env, "foo=error,bar=debug")
	defer func() {
		os.Setenv(Env, oldEnv)
	}()
	b := leveledCache{}
	require.True(t, b.Enabled("foo", Error))
	require.False(t, b.Enabled("foo", Info))
	require.False(t, b.Enabled("foo", Debug))
	require.True(t, b.Enabled("bar", Error))
	require.True(t, b.Enabled("bar", Info))
	require.True(t, b.Enabled("bar", Debug))

	// Log levels can be read from a custom environment variable.
	// The custom one takes precedence over the default one.
	os.Setenv("CUSTOM_LOG_LEVELS", "foo=debug")
	defer func() {
		os.Unsetenv("CUSTOM_LOG_LEVELS")
	}()
	c := leveledCache{Env: "CUSTOM_LOG_LEVELS"}
	require.True(t, c.Enabled("foo", Error))
	require.True(t, c.Enabled("foo", Info))
	require.True(t, c.Enabled("foo", Debug))
	require.True(t, c.Enabled("bar", Error))
	require.True(t, c.Enabled("bar", Info))
	require.False(t, c.Enabled("bar", Debug))
}
