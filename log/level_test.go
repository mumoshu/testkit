package log_test

import (
	"testing"

	"github.com/mumoshu/testkit/log"
	"github.com/stretchr/testify/require"
)

func TestLevel(t *testing.T) {
	require.Less(t, log.Error, log.Info)
	require.Less(t, log.Info, log.Debug)

	require.Equal(t, "error", log.Error.String())
	require.Equal(t, "info", log.Info.String())
	require.Equal(t, "debug", log.Debug.String())
	require.Equal(t, "(unknown log level: 3)", log.Level(3).String())
}
