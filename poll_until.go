package testkit

import (
	"testing"
	"time"
)

func PollUntil(t *testing.T, condition func() bool, timeout time.Duration) {
	t.Helper()

	start := time.Now()
	for {
		if condition() {
			return
		}
		if time.Since(start) > timeout {
			t.Fatal("timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
