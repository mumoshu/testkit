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
}

func NewKubectl(kubeconfigPath string) *Kubectl {
	return &Kubectl{
		KubeconfigPath: kubeconfigPath,
	}
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
	return string(r), err
}
