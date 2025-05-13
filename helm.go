package testkit

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type Helm struct {
	// KubeconfigPath is the path to the kubeconfig file.
	KubeconfigPath string
}

func NewHelm(kubeconfigPath string) *Helm {
	return &Helm{
		KubeconfigPath: kubeconfigPath,
	}
}

type HelmConfig struct {
	ExtraArgs []string
	Namespace string
	Values    map[string]interface{}
	Version   string
}

type HelmOption func(*HelmConfig)

func (k *Helm) UpgradeOrInstall(t *testing.T, releaseName, chartPath string, opts ...HelmOption) {
	t.Helper()

	var c HelmConfig

	for _, o := range opts {
		o(&c)
	}

	var args []string

	if c.Namespace != "" {
		args = append(args, "--namespace", c.Namespace)
	}

	args = append(args, "upgrade", "--install", releaseName, chartPath)
	args = append(args, "--wait", "--timeout", "5m", "--create-namespace", "--atomic", "--debug")

	if c.Version != "" {
		args = append(args, "--version", c.Version)
	}

	if c.Values != nil {
		args = append(args, "--values", "values.yaml")

		f, err := os.Create("values.yaml")
		require.NoError(t, err)

		err = writeHelmValuesYAMLFile(f, c.Values)
		require.NoError(t, err)

		require.NoError(t, f.Sync())
		require.NoError(t, f.Close())

		defer func() {
			err := os.Remove("values.yaml")
			require.NoError(t, err)
		}()
	}

	args = append(args, c.ExtraArgs...)

	_, err := k.capture(args...)
	require.NoError(t, err)
}

func (k *Helm) AddRepo(t *testing.T, repoName, repoURL string) {
	t.Helper()

	_, err := k.capture("repo", "add", repoName, repoURL)
	require.NoError(t, err)
}

func (k *Helm) UpdateRepo(t *testing.T, repoName string) {
	t.Helper()

	_, err := k.capture("repo", "update", repoName)
	require.NoError(t, err)
}

func (k *Helm) Capture(t *testing.T, args ...string) string {
	t.Helper()

	r, err := k.capture(args...)
	require.NoError(t, err)
	return r
}

func writeHelmValuesYAMLFile(f *os.File, values map[string]interface{}) error {
	enc := yaml.NewEncoder(f)
	defer enc.Close()

	if err := enc.Encode(values); err != nil {
		return err
	}

	return nil
}

func (k *Helm) capture(args ...string) (string, error) {
	c := exec.Command("helm", args...)
	c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", k.KubeconfigPath))

	r, err := c.CombinedOutput()
	if err != nil {
		errWithOutput := fmt.Errorf("error running helm command: %w, output: %s", err, string(r))
		return string(r), errWithOutput
	}

	return string(r), nil
}
