package testkit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type KindProvider struct {
	kindBin string

	// clusterNames is a list of cluster names that have been created.
	clusterNames map[string]struct{}

	// kubeconfigDir is the directory where the kubeconfig files are stored.
	kubeconfigDir string
}

var _ Provider = &KindProvider{}
var _ KubernetesClusterProvider = &KindProvider{}

func (p *KindProvider) Setup() error {
	const (
		kindBin = "kind"
	)

	bin, err := exec.LookPath(kindBin)
	if err != nil {
		return fmt.Errorf("unable to find %s binary: %v", kindBin, err)
	}

	p.kindBin = bin
	p.kubeconfigDir = filepath.Join(os.TempDir(), "testkit_kind_kubeconfigs")
	p.clusterNames = make(map[string]struct{})

	return nil
}

func (p *KindProvider) Cleanup() error {
	for clusterName := range p.clusterNames {
		_, err := p.capture(p.clusterKubeconfigPath(clusterName), "delete", "cluster", "--name", clusterName)
		if err != nil {
			return fmt.Errorf("unable to delete cluster %s: %v", clusterName, err)
		}
	}

	return nil
}

func (p *KindProvider) clusterKubeconfigPath(clusterName string) string {
	return filepath.Join(p.kubeconfigDir, fmt.Sprintf("%s.kubeconfig", clusterName))
}

func (p *KindProvider) capture(kubeconfigPath string, args ...string) (string, error) {
	c := exec.Command(p.kindBin, args...)
	c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath))

	r, err := c.CombinedOutput()
	return string(r), err
}

func (p *KindProvider) GetKubernetesCluster(opts ...KubernetesClusterOption) (*KubernetesCluster, error) {
	var conf KubernetesClusterConfig

	for _, opt := range opts {
		opt(&conf)
	}

	clusterName := "testkit-"
	if conf.ID != "" {
		clusterName += conf.ID + "-"
	}

	for cn := range p.clusterNames {
		if strings.HasPrefix(cn, clusterName) {
			kubeconfigPath := p.clusterKubeconfigPath(cn)

			_, err := p.capture(kubeconfigPath, "export", "kubeconfig", "--name", cn)
			if err != nil {
				return nil, fmt.Errorf("unable to export kubeconfig for cluster %s: %v", cn, err)
			}

			return &KubernetesCluster{
				KubeconfigPath: kubeconfigPath,
			}, nil
		}
	}

	var unmanagedAvailableClusterNames []string
	{
		r, err := p.capture("", "get", "clusters")
		if err != nil {
			return nil, fmt.Errorf("unable to get clusters: %v", err)
		}

		for _, line := range strings.Split(r, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			unmanagedAvailableClusterNames = append(unmanagedAvailableClusterNames, line)
		}
	}

	for _, cn := range unmanagedAvailableClusterNames {
		if strings.HasPrefix(cn, clusterName) {
			kubeconfigPath := p.clusterKubeconfigPath(cn)

			_, err := p.capture(kubeconfigPath, "export", "kubeconfig", "--name", cn)
			if err != nil {
				return nil, fmt.Errorf("unable to export kubeconfig for cluster %s: %v", cn, err)
			}

			// We don't want to delete the cluster when we're done with it.
			// That's because we didn't create it.
			// p.clusterNames[clusterName] = struct{}{}

			return &KubernetesCluster{
				KubeconfigPath: kubeconfigPath,
			}, nil
		}
	}

	clusterName += randString(4)

	kubeconfigPath := p.clusterKubeconfigPath(clusterName)

	_, err := p.capture(kubeconfigPath, "create", "cluster", "--name", clusterName)
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster %s: %v", clusterName, err)
	}

	p.clusterNames[clusterName] = struct{}{}

	return &KubernetesCluster{
		KubeconfigPath: kubeconfigPath,
	}, nil
}
