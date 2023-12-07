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

func (tk *TestKit) KubernetesNamespace(t *testing.T, opts ...KubernetesNamespaceOption) *KubernetesNamespace {
	t.Helper()

	var cp KubernetesNamespaceProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(KubernetesNamespaceProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no KubernetesNamespaceProvider found")
	}

	ns, err := cp.KubernetesNamespace(opts...)
	if err != nil {
		t.Fatalf("unable to get s3 bucket: %v", err)
	}

	return ns
}
