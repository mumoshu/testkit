package testkit

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
)

type KubectlProvider struct {
	// DefaultKubeconfigPath is the path to the kubeconfig file.
	DefaultKubeconfigPath string

	kubeconfigToResources map[string]kubectlResources
}

type kubectlResources struct {
	namespaces map[string]struct{}
}

func (p *kubectlResources) addNamespace(name string) {
	if p.namespaces == nil {
		p.namespaces = make(map[string]struct{})
	}

	p.namespaces[name] = struct{}{}
}

func (p *kubectlResources) getNamespaces() []string {
	var namespaces []string
	for namespace := range p.namespaces {
		namespaces = append(namespaces, namespace)
	}

	return namespaces
}

var _ Provider = &KubectlProvider{}
var _ KubernetesNamespaceProvider = &KubectlProvider{}

func (p *KubectlProvider) Setup() error {
	if p.DefaultKubeconfigPath != "" {
		_, err := os.Stat(p.DefaultKubeconfigPath)
		if err != nil {
			return fmt.Errorf("unable to stat kubeconfig file: %v", err)
		}
	}

	p.kubeconfigToResources = make(map[string]kubectlResources)

	return nil
}

func (p *KubectlProvider) Cleanup() error {
	for kubeconfigPath, resources := range p.kubeconfigToResources {
		for _, ns := range resources.getNamespaces() {
			kubectl := NewKubectl(kubeconfigPath)
			_, err := kubectl.capture("delete", "namespace", ns)
			if err != nil {
				return fmt.Errorf("unable to delete namespace %s/%s: %v", kubeconfigPath, ns, err)
			}
		}
	}

	return nil
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = letters[b[i]%byte(len(letters))]
	}

	return string(b)
}

func (p *KubectlProvider) KubernetesNamespace(opts ...KubernetesNamespaceOption) (*KubernetesNamespace, error) {
	config := &KubernetesNamespaceConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.KubeconfigPath == "" {
		config.KubeconfigPath = p.DefaultKubeconfigPath
	}

	resources, ok := p.kubeconfigToResources[config.KubeconfigPath]
	if !ok {
		resources = kubectlResources{}
		p.kubeconfigToResources[config.KubeconfigPath] = resources
	}

	// nsName can be empty, in which case we'll use the first namespace, if any.
	// If there are no namespaces, we'll create one.
	nsName := "testkit-"
	if config.ID != "" {
		nsName += config.ID + "-"
	}

	var foundNsName string
	for _, ns := range resources.getNamespaces() {
		if strings.HasPrefix(ns, nsName) {
			foundNsName = ns
			break
		}
	}

	if foundNsName != "" {
		return &KubernetesNamespace{
			Name: foundNsName,
		}, nil
	}

	nsName += randString(5)

	kubectl := NewKubectl(config.KubeconfigPath)

	_, err := kubectl.capture("create", "namespace", nsName, "--kubeconfig", config.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	resources.addNamespace(nsName)

	return &KubernetesNamespace{
		Name: nsName,
	}, nil
}
