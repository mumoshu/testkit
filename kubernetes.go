package testkit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type Kubernetes struct {
	kubectl *Kubectl
}

func NewKubernetes(kubeconfigPath string) *Kubernetes {
	return &Kubernetes{
		kubectl: NewKubectl(kubeconfigPath),
	}
}

// ListReadyNodeNames returns the names of the nodes that are ready.
// It's useful when you want to test your app that adds or removes nodes.
//
// The readiness of a node is determined by the status of the node.
// A node is ready if the status of the node has a condition of type "Ready" and the status is "True".
func (k *Kubernetes) ListReadyNodeNames(t *testing.T) []string {
	t.Helper()

	nodes := k.GetNodes(t)

	var readyNodes []string

	for _, node := range nodes {
		if node.IsReady() {
			readyNodes = append(readyNodes, node.Metadata.Name)
		}
	}

	return readyNodes
}

// GetNodes returns the nodes in the cluster.
// It's useful when you want to test your app that adds, removes, updates, or refers to nodes.
func (k *Kubernetes) GetNodes(t *testing.T) []KubernetesNode {
	t.Helper()

	var cmd []string

	cmd = append(cmd, "get", "nodes", "-o", "json")

	out := k.capture(t, cmd...)

	var result struct {
		Items []KubernetesNode `json:"items"`
	}

	err := json.Unmarshal([]byte(out), &result)
	require.NoError(t, err)

	return result.Items
}

type KubernetesNode struct {
	Metadata kubernetesMetadata   `json:"metadata"`
	Status   kubernetesNodeStatus `json:"status"`
}

func (n *KubernetesNode) IsReady() bool {
	for _, c := range n.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			return true
		}
	}

	return false
}

type kubernetesMetadata struct {
	Name string `json:"name"`
}

type kubernetesNodeStatus struct {
	Conditions []kubernetesNodeCondition `json:"conditions"`
}

type kubernetesNodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (k *Kubernetes) capture(t *testing.T, args ...string) string {
	t.Helper()

	return k.kubectl.Capture(t, args...)
}
