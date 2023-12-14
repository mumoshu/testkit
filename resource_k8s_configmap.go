package testkit

import "testing"

// KubernetesConfigMapProvider is a provider that can provision an Kubernetes ConfigMap.
// Any provider that can create a ConfigMap should implement this interface.
type KubernetesConfigMapProvider interface {
	KubernetesConfigMap(opts ...KubernetesConfigMapOption) (*KubernetesConfigMap, error)
}

type KubernetesConfigMap struct {
	Namespace string
	Name      string
}

type KubernetesConfigMapConfig struct {
	ID             string
	Namespace      string
	KubeconfigPath string
}

type KubernetesConfigMapOption func(*KubernetesConfigMapConfig)

func KubernetesConfigMapKubeconfigPath(path string) KubernetesConfigMapOption {
	return func(c *KubernetesConfigMapConfig) {
		c.KubeconfigPath = path
	}
}

func KubernetesConfigMapNamespace(namespace string) KubernetesConfigMapOption {
	return func(c *KubernetesConfigMapConfig) {
		c.Namespace = namespace
	}
}

// KubernetesConfigMap returns a KubernetesConfigMap.
// It does so by iterating over the available providers and calling the KubernetesConfigMap method on each provider.
// If no provider implements KubernetesConfigMap, it fails the test.
// If multiple providers implement KubernetesConfigMap, it returns the first successful one.
// If multiple providers implement KubernetesConfigMap and all of them fail, it fails the test.
func (tk *TestKit) KubernetesConfigMap(t *testing.T, opts ...KubernetesConfigMapOption) *KubernetesConfigMap {
	t.Helper()

	var cp KubernetesConfigMapProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(KubernetesConfigMapProvider)
		if ok {
			ns, err := cp.KubernetesConfigMap(opts...)
			if err != nil {
				t.Logf("unable to get configmap: %v", err)
				continue
			}

			return ns
		}
	}

	if cp == nil {
		t.Fatal("no KubernetesConfigMapProvider found")
	}

	return nil
}
