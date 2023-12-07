package testkit

type EKSCTLProvider struct {
	// ConfigPath is the path to the eksctl config file.
	ConfigPath string
}

var _ Provider = &EKSCTLProvider{}
var _ EKSClusterProvider = &EKSCTLProvider{}

func (p *EKSCTLProvider) GetEKSCluster(opts ...EKSClusterOption) (*EKSCluster, error) {
	return nil, nil
}

func (p *EKSCTLProvider) Setup() error {
	return nil
}

func (p *EKSCTLProvider) Cleanup() error {
	return nil
}
