package testkit

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

type Kubectl struct {
	// KubeconfigPath is the path to the kubeconfig file.
	KubeconfigPath string

	// LogError controls whether to log the error returned by kubectl.
	// Currently, this is respected that doesn't return the error,
	// like the `Failed` method.
	LogError bool
}

func NewKubectl(kubeconfigPath string) *Kubectl {
	return &Kubectl{
		KubeconfigPath: kubeconfigPath,
	}
}

// Failed runs kubectl with the given args and returns true if it fails.
// This is useful for e.g. asserting that a resource does not exist.
func (k *Kubectl) Failed(t *testing.T, args ...string) bool {
	t.Helper()

	_, err := k.capture(args...)

	if k.LogError {
		t.Log(err)
	}

	return err != nil
}

func (k *Kubectl) Capture(t *testing.T, args ...string) string {
	t.Helper()

	r, err := k.capture(args...)
	require.NoError(t, err)
	return r
}

func (k *Kubectl) capture(args ...string) (string, error) {
	c := exec.Command("kubectl", args...)
	c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", k.KubeconfigPath))

	r, err := c.CombinedOutput()
	if err != nil {
		errWithOutput := fmt.Errorf("error running kubectl command: %w, output: %s", err, string(r))
		return string(r), errWithOutput
	}
	return string(r), nil
}
