package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type leveledCache struct {
	// Env is the environment variable that contains the log entries.
	// If not set, it defaults to log.Env, which is "TESTKIT_LOG" by default.
	Env string

	// Level is the default log level, which is log.Info, used
	// when the log entry is not found in the cache.
	Level Level

	entries map[string]Level

	mu sync.Mutex
}

// Enabled returns true if the given log entry is enabled.
//
// The log entry is enabled if the TESTKIT_LOG environment variable contains the entry.
// Each entry is separated by a comma, and in the form of "name=level".
// The name is the name of the component that is logging, and the level is the log level.
// The level can be one of the following: "debug", "info", "warn", "error".
//
// For example, "kind=debug" enables debug logging for the kind provider.
// The component that is logging should check if the log entry is enabled before logging.
//
// That said, every component's logging code should look like this:
//
//	if l.Enabled("kind=debug") {
//	    l.Logf("debug message")
//	}
//
//	if l.Enabled("terraform=info") {
//	    l.Logf("info message")
//	}
//
// Any Logf call that is not wrapped in an Enabled call is considered a bug.
//
// The debug log is only printed if the TESTKIT_LOG environment variable contains any of the following entries:
// - "kind=debug"
// - "debug"
//
// The info log is only printed if the TESTKIT_LOG environment variable contains any of the following entries:
// - "terraform=info"
// - "info"
// - "debug"
//
// As said previously, TESTKIT_LOG is a comma-separated list of log entries.
// TESTKIT_LOG=kind=debug,terraform=info is valid, and enables debug logging for the kind provider and info logging for the terraform provider.
func (c *leveledCache) Enabled(name string, level Level) bool {
	c.ensureLoaded()

	if l, ok := c.entries[name]; ok {
		return l >= level
	}

	return c.Level >= level
}

func (c *leveledCache) ensureLoaded() {
	if c.entries != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := levelCacheFromEnv(c.Env)
	if err != nil {
		panic(err)
	}

	c.entries = entries

	c.Level = entries[""]
}

func levelCacheFromEnv(env string) (map[string]Level, error) {
	if env == "" {
		env = Env
	}

	entries := make(map[string]Level)

	logEntries := strings.Split(os.Getenv(env), ",")
	for _, enabledEntry := range logEntries {
		name, level, err := parseLevel(enabledEntry)
		if err != nil {
			return nil, fmt.Errorf("unable to parse log entry %q: %v", enabledEntry, err)
		}

		entries[name] = level
	}

	return entries, nil
}

func parseLevel(enabledEntry string) (string, Level, error) {
	var (
		name  string
		level Level
	)

	splits := strings.Split(enabledEntry, "=")
	switch len(splits) {
	case 1:
		switch splits[0] {
		case "debug":
			level = Debug
		case "info":
			level = Info
		case "error":
			level = Error
		case "":
			level = Info
		default:
			return "", 0, fmt.Errorf("invalid log level for %s: %s", name, splits[0])
		}
	case 2:
		name = splits[0]
		switch splits[1] {
		case "debug":
			level = Debug
		case "info":
			level = Info
		case "error":
			level = Error
		default:
			return "", 0, fmt.Errorf("invalid log level for %s: %s", name, splits[1])
		}
	default:
		return "", 0, fmt.Errorf("invalid log entry: %s", enabledEntry)
	}

	return name, level, nil
}
