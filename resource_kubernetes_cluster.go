package testkit

import "testing"

type KubernetesCluster struct {
	// KubeconfigPath is the path to the kubeconfig file.
	KubeconfigPath string
}

type KubernetesClusterConfig struct {
	ID string
}

type KubernetesClusterOption func(*KubernetesClusterConfig)

type KubernetesClusterProvider interface {
	GetKubernetesCluster(...KubernetesClusterOption) (*KubernetesCluster, error)
}

func (tk *TestKit) KubernetesCluster(t *testing.T, opts ...KubernetesClusterOption) *KubernetesCluster {
	var cp KubernetesClusterProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(KubernetesClusterProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no KubernetesClusterProvider found")
	}

	kc, err := cp.GetKubernetesCluster(opts...)
	if err != nil {
		t.Fatalf("unable to get kubernetes cluster: %v", err)
	}

	return kc
}
