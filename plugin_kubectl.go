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

	kubeconfigToResources map[string]*kubectlResources
}

type kubectlResources struct {
	configmaps map[string]map[string]struct{}
	namespaces map[string]struct{}
}

func (p *kubectlResources) addConfigMap(ns, name string) {
	if p.configmaps == nil {
		p.configmaps = make(map[string]map[string]struct{})
	}

	if p.configmaps[ns] == nil {
		p.configmaps[ns] = make(map[string]struct{})
	}

	p.configmaps[ns][name] = struct{}{}
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

	p.kubeconfigToResources = make(map[string]*kubectlResources)

	return nil
}

func (p *KubectlProvider) Cleanup() error {
	for kubeconfigPath, resources := range p.kubeconfigToResources {
		kubectl := NewKubectl(kubeconfigPath)

		for ns, cms := range resources.configmaps {
			for cm := range cms {
				_, err := kubectl.capture("delete", "configmap", cm, "--namespace", ns)
				if err != nil {
					return fmt.Errorf("unable to delete configmap %s/%s: %v", kubeconfigPath, cm, err)
				}
			}
		}

		for _, ns := range resources.getNamespaces() {
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

func (p *KubectlProvider) KubernetesConfigMap(opts ...KubernetesConfigMapOption) (*KubernetesConfigMap, error) {
	config := &KubernetesConfigMapConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.KubeconfigPath == "" {
		config.KubeconfigPath = p.DefaultKubeconfigPath
	}

	resources, ok := p.kubeconfigToResources[config.KubeconfigPath]
	if !ok {
		resources = &kubectlResources{}
		p.kubeconfigToResources[config.KubeconfigPath] = resources
	}

	if resources.configmaps == nil {
		resources.configmaps = make(map[string]map[string]struct{})
	}

	nsName := config.Namespace
	if nsName == "" {
		nsName = "default"
	}

	// cmName can be empty, in which case we'll use the first namespace, if any.
	// If there are no namespaces, we'll create one.
	cmName := "testkit-"
	if config.ID != "" {
		cmName += config.ID + "-"
	}

	cms := resources.configmaps[nsName]
	if cms == nil {
		cms = make(map[string]struct{})
		resources.configmaps[nsName] = cms
	}

	var foundCMName string
	for cm := range cms {
		if strings.HasPrefix(cm, cmName) {
			foundCMName = cm
			break
		}
	}

	if foundCMName != "" {
		return &KubernetesConfigMap{
			Namespace: nsName,
			Name:      foundCMName,
		}, nil
	}

	cmName += randString(5)

	kubectl := NewKubectl(config.KubeconfigPath)

	_, err := kubectl.capture("create", "configmap", cmName, "--kubeconfig", config.KubeconfigPath, "--namespace", nsName)
	if err != nil {
		return nil, err
	}

	resources.addConfigMap(nsName, cmName)

	return &KubernetesConfigMap{
		Namespace: nsName,
		Name:      cmName,
	}, nil
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
		resources = &kubectlResources{}
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
