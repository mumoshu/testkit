package testkit

import "testing"

// EKSClusterProvider is a provider that can provision an EKS cluster.
// Any provider that can create an EKS cluster should implement this interface.
type EKSClusterProvider interface {
	// GetEKSCluster returns an EKS cluster.
	GetEKSCluster(opts ...EKSClusterOption) (*EKSCluster, error)
}

// EKSCluster is an EKS cluster that is
// provided by the TestKit.
// It doesn't necessarily be created by the TestKit,
// but it can be created by other tools,
// such as eksctl.
type EKSCluster struct {
	Endpoint       string
	KubeconfigPath string
}

// EKSClusterOptions is the options for creating an EKS cluster.
// The zero value is a valid value.
type EKSClusterConfig struct {
	// The ID is the local name of the cluster in the test,
	// which is used by the provider implmentation to identify
	// the cluster in the test.
	// The ID is usually not the same as the name of the cluster.
	ID string
}

type EKSClusterOption func(*EKSClusterConfig)

// EKSCluster creates an EKS cluster.
func (tk *TestKit) EKSCluster(t *testing.T, opts ...EKSClusterOption) *EKSCluster {
	t.Helper()

	var cp EKSClusterProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(EKSClusterProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no EKSClusterProvider found")
	}

	eksCluster, err := cp.GetEKSCluster(opts...)
	if err != nil {
		t.Fatalf("unable to get eks cluster: %v", err)
	}

	return eksCluster
}
