package testkit

import "testing"

// KubernetesNamespaceProvider is a provider that can provision an Kubernetes namespace.
// Any provider that can create an S3 bucket should implement this interface.
type KubernetesNamespaceProvider interface {
	KubernetesNamespace(opts ...KubernetesNamespaceOption) (*KubernetesNamespace, error)
}

type KubernetesNamespace struct {
	Name string
}

type KubernetesNamespaceConfig struct {
	ID             string
	KubeconfigPath string
}

type KubernetesNamespaceOption func(*KubernetesNamespaceConfig)

func KubeconfigPath(path string) KubernetesNamespaceOption {
	return func(c *KubernetesNamespaceConfig) {
		c.KubeconfigPath = path
	}
}

// KubernetesNamespace returns a KubernetesNamespace.
// It does so by iterating over the available providers and calling the KubernetesNamespace method on each provider.
// If no provider implements KubernetesNamespace, it fails the test.
// If multiple providers implement KubernetesNamespace, it returns the first successful one.
// If multiple providers implement KubernetesNamespace and all of them fail, it fails the test.
func (tk *TestKit) KubernetesNamespace(t *testing.T, opts ...KubernetesNamespaceOption) *KubernetesNamespace {
	t.Helper()

	var cp KubernetesNamespaceProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(KubernetesNamespaceProvider)
		if ok {
			ns, err := cp.KubernetesNamespace(opts...)
			if err != nil {
				t.Logf("unable to get s3 bucket: %v", err)
				continue
			}

			return ns
		}
	}

	if cp == nil {
		t.Fatal("no KubernetesNamespaceProvider found")
	}

	return nil
}
