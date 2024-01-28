package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type TerraformProvider struct {
	// WorkspacePath is the path to the Terraform workspace.
	WorkspacePath string
	// Vars is the map of Terraform variables.
	Vars map[string]string

	// KubeconfigDir is the directory where the kubeconfig files for EKS clusters are stored.
	KubeconfigDir string

	tfShowJSONBytes []byte
}

var _ S3BucketProvider = &TerraformProvider{}
var _ EKSClusterProvider = &TerraformProvider{}
var _ KubernetesClusterProvider = &TerraformProvider{}
var _ Provider = &TerraformProvider{}

func (p *TerraformProvider) GetEKSCluster(opts ...EKSClusterOption) (*EKSCluster, error) {
	if p.KubeconfigDir == "" {
		return nil, fmt.Errorf("kubeconfigDir is not set")
	}

	resource, err := p.getEKSClusterResource()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := p.generateKubeconfigFile(resource.tfEKSClusterValues)
	if err != nil {
		return nil, err
	}

	return &EKSCluster{
		Endpoint:       resource.tfEKSClusterValues.Endpoint,
		KubeconfigPath: kubeconfigPath,
	}, nil
}

func (p *TerraformProvider) GetKubernetesCluster(opts ...KubernetesClusterOption) (*KubernetesCluster, error) {
	var conf KubernetesClusterConfig

	for _, opt := range opts {
		opt(&conf)
	}

	resource, err := p.getEKSClusterResource()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := p.generateKubeconfigFile(resource.tfEKSClusterValues)
	if err != nil {
		return nil, err
	}

	return &KubernetesCluster{
		KubeconfigPath: kubeconfigPath,
	}, nil
}

func (p *TerraformProvider) getEKSClusterResource() (tfResource, error) {
	resources, err := p.readEKSClusterResources(bytes.NewReader(p.tfShowJSONBytes))
	if err != nil {
		return tfResource{}, err
	}

	for _, resource := range resources {
		if resource.tfEKSClusterValues != nil {
			return resource, nil
		}
	}

	return tfResource{}, fmt.Errorf("unable to find EKS cluster resource")
}

func (p *TerraformProvider) generateKubeconfigFile(values *tfEKSClusterValues) (string, error) {
	kubeconfigPath := filepath.Join(p.KubeconfigDir, "tfeks_"+values.ID+".kubeconfig")

	kubeconfig := &Kubeconfig{}
	kubeconfig.APIVersion = "v1"
	kubeconfig.Kind = "Config"
	kubeconfigClusterUserContextName := "testkit_" + values.ID
	var (
		cluster KubeconfigCluster
		context KubeconfigContext
		user    KubeconfigUser
	)
	cluster.Name = kubeconfigClusterUserContextName
	cluster.Cluster.Server = values.Endpoint
	cluster.Cluster.CertificateAuthorityData = values.CertificateAuthority[0].Data
	context.Name = kubeconfigClusterUserContextName
	context.Context.Cluster = kubeconfigClusterUserContextName
	context.Context.User = kubeconfigClusterUserContextName
	user.Name = kubeconfigClusterUserContextName
	user.User.Exec.APIVersion = "client.authentication.k8s.io/v1beta1"
	user.User.Exec.Command = "aws"
	region := strings.Split(values.ARN, ":")[3]
	user.User.Exec.Args = []string{
		"--region",
		region,
		"eks",
		"get-token",
		"--cluster-name",
		values.ID,
		"--output",
		"json",
	}
	kubeconfig.Clusters = []KubeconfigCluster{cluster}
	kubeconfig.Contexts = []KubeconfigContext{context}
	kubeconfig.Users = []KubeconfigUser{user}
	kubeconfig.CurrentContext = kubeconfigClusterUserContextName

	kubeconfigInYaml, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("unable to marshal kubeconfig: %v", err)
	}

	if err := os.MkdirAll(p.KubeconfigDir, 0755); err != nil {
		return "", fmt.Errorf("unable to create kubeconfig directory %q: %v", p.KubeconfigDir, err)
	}

	// 0600 instead of 0644 for security, more concretely to avoid the following warnings:
	//   WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /path/to/the/kubeconfig/file
	//   WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /path/to/the/kubeconfig/file
	if err := os.WriteFile(kubeconfigPath, kubeconfigInYaml, 0600); err != nil {
		return "", fmt.Errorf("unable to write kubeconfig file: %v", err)
	}

	return kubeconfigPath, nil
}

type Kubeconfig struct {
	APIVersion     string              `yaml:"apiVersion"`
	Clusters       []KubeconfigCluster `yaml:"clusters"`
	Contexts       []KubeconfigContext `yaml:"contexts"`
	CurrentContext string              `yaml:"current-context"`
	Kind           string              `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []KubeconfigUser `yaml:"users"`
}

type KubeconfigCluster struct {
	Cluster struct {
		Server                   string `yaml:"server"`
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
	} `yaml:"cluster"`
	Name string `yaml:"name"`
}

type KubeconfigContext struct {
	Context struct {
		Cluster string `yaml:"cluster"`
		User    string `yaml:"user"`
	} `yaml:"context"`
	Name string `yaml:"name"`
}

type KubeconfigUser struct {
	Name string `yaml:"name"`
	User struct {
		Exec struct {
			APIVersion string   `yaml:"apiVersion"`
			Command    string   `yaml:"command"`
			Args       []string `yaml:"args"`
		} `yaml:"exec"`
	} `yaml:"user"`
}

func (p *TerraformProvider) GetS3Bucket(opts ...S3BucketOption) (*S3Bucket, error) {
	resources, err := p.readS3BucketResources(bytes.NewReader(p.tfShowJSONBytes))
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		if resource.S3BucketValues == nil {
			continue
		}

		return &S3Bucket{
			Name:   resource.S3BucketValues.Bucket,
			Region: resource.S3BucketValues.Region,
		}, nil
	}

	return nil, fmt.Errorf("unable to find S3 bucket")
}

func (p *TerraformProvider) GetECRImageRepository(opts ...ECRImageRepositoryOption) (*ECRImageRepository, error) {
	resources, err := p.readECRImageRepository(bytes.NewReader(p.tfShowJSONBytes))
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		if resource.tfECRImageRepoValues == nil {
			continue
		}

		return &ECRImageRepository{
			ID:            resource.tfECRImageRepoValues.ID,
			ARN:           resource.tfECRImageRepoValues.ARN,
			RepositoryURL: resource.tfECRImageRepoValues.RepositoryURL,
			RegistryID:    resource.tfECRImageRepoValues.RegistryID,
		}, nil
	}

	return nil, fmt.Errorf("unable to find ECR image repository")
}

type tfResource struct {
	Address string `json:"address"`
	// Mode can be e.g. "managed" or "data".
	Mode string `json:"mode"`
	// Type can be e.g.:
	// - aws_s3_bucket
	// - aws_iam_role
	// - aws_vpc
	// - aws_eks_cluster
	Type string `json:"type"`
	// Name is the name of the Terraform resource.
	Name string `json:"name"`
	// ProviderName is the name of the Terraform provider.
	// E.g. "registry.terraform.io/hashicorp/aws".
	ProviderName string          `json:"provider_name"`
	Values       json.RawMessage `json:"values"`

	S3BucketValues       *tfS3BucketValues           `json:"-"`
	tfEKSClusterValues   *tfEKSClusterValues         `json:"-"`
	tfECRImageRepoValues *tfECRImageRepositoryValues `json:"-"`
}

type tfS3BucketValues struct {
	Bucket string `json:"bucket"`
	// ID and Bucket have the same value.
	ID                       string `json:"id"`
	BucketDomainName         string `json:"bucket_domain_name"`
	BucketRegionalDomainName string `json:"bucket_regional_domain_name"`
	Region                   string `json:"region"`
}

type tfEKSClusterValues struct {
	// ID is the name of the EKS cluster.
	ID                   string                             `json:"name"`
	Endpoint             string                             `json:"endpoint"`
	ARN                  string                             `json:"arn"`
	CertificateAuthority []tfEKSClusterCertificateAuthority `json:"certificate_authority"`
}

type tfECRImageRepositoryValues struct {
	// ID is the name of the ECR image repository.
	// In case the ARN is:
	// 	arn:aws:ecr:${REGION}:${ACCOUNT_ID}:repository/testkit-imagerep
	// the ID is:
	// 	testkit-imagerep
	ID string `json:"id"`
	// ARN is the Amazon Resource Name of the ECR image repository.
	// In case the name of the repository is "testkit-imagerep",
	// the ARN is:
	// 	arn:aws:ecr:${REGION}:${ACCOUNT_ID}:repository/testkit-imagerep
	ARN string `json:"arn"`
	// RepositoryURL is the URL of the ECR image repository.
	// In case the name of the repository is "testkit-imagerep",
	// the URL is:
	// 	${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/testkit-imagerep
	RepositoryURL string `json:"repository_url"`
	// RegistryID is the ID of the registry.
	// It's the same as the account ID.
	RegistryID string `json:"registry_id"`
}

type tfEKSClusterCertificateAuthority struct {
	Data string `json:"data"`
}

type tfShowOutput struct {
	Values tfShowValues `json:"values"`
}

type tfShowValues struct {
	RootModule tfShowRootModule `json:"root_module"`
}

type tfShowRootModule struct {
	Resources []tfResource `json:"resources"`
}

func (p *TerraformProvider) captureTerraformShowJSON() ([]byte, error) {
	output, err := p.runTerraformCommand("show", "-json")
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (p *TerraformProvider) runTerraformCommand(args ...string) ([]byte, error) {
	var argsWithVars []string

	argsWithVars = append(argsWithVars, args...)

	for k, v := range p.Vars {
		argsWithVars = append(argsWithVars, "-var", k+"="+v)
	}

	return p.runTerraformCommandNoVars(argsWithVars...)
}

func (p *TerraformProvider) runTerraformCommandNoVars(args ...string) ([]byte, error) {
	c := exec.Command("terraform", args...)
	c.Dir = p.WorkspacePath

	r, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unable to run terraform command: %v\n%s", err, string(r))
	}

	return r, nil
}

func (p *TerraformProvider) readEKSClusterResources(r io.Reader) ([]tfResource, error) {
	rs, err := p.readResourcesOfType(r, "aws_eks_cluster")
	if err != nil {
		return nil, err
	}

	var filteredResources []tfResource

	for _, r := range rs {
		var values tfEKSClusterValues
		if err := json.Unmarshal(r.Values, &values); err != nil {
			return nil, err
		}

		r.tfEKSClusterValues = &values

		filteredResources = append(filteredResources, r)
	}

	return filteredResources, nil
}

func (p *TerraformProvider) readS3BucketResources(r io.Reader) ([]tfResource, error) {
	rs, err := p.readResourcesOfType(r, "aws_s3_bucket")
	if err != nil {
		return nil, fmt.Errorf("unable to read S3 bucket resources: %v", err)
	}

	var filteredResources []tfResource

	for _, r := range rs {
		var values tfS3BucketValues
		if err := json.Unmarshal(r.Values, &values); err != nil {
			return nil, fmt.Errorf("unable to unmarshal tfS3BucketValues: %v", err)
		}

		r.S3BucketValues = &values

		filteredResources = append(filteredResources, r)
	}

	return filteredResources, nil
}

func (p *TerraformProvider) readECRImageRepository(r io.Reader) ([]tfResource, error) {
	rs, err := p.readResourcesOfType(r, "aws_ecr_repository")
	if err != nil {
		return nil, err
	}

	var filteredResources []tfResource

	for _, r := range rs {
		var values tfECRImageRepositoryValues
		if err := json.Unmarshal(r.Values, &values); err != nil {
			return nil, err
		}

		r.tfECRImageRepoValues = &values

		filteredResources = append(filteredResources, r)
	}

	return filteredResources, nil
}

func (p *TerraformProvider) readResourcesOfType(r io.Reader, resourceType string) ([]tfResource, error) {
	resources, err := p.readResources(r)
	if err != nil {
		return nil, err
	}

	var filteredResources []tfResource
	for _, resource := range resources {
		if resource.Type == resourceType {
			filteredResources = append(filteredResources, resource)
		}
	}

	return filteredResources, nil
}

func (p *TerraformProvider) readResources(r io.Reader) ([]tfResource, error) {
	var output tfShowOutput
	if err := json.NewDecoder(r).Decode(&output); err != nil {
		return nil, err
	}

	return output.Values.RootModule.Resources, nil
}

func (p *TerraformProvider) Setup() error {
	if p.KubeconfigDir == "" {
		p.KubeconfigDir = filepath.Join(os.TempDir(), "testkit_terraform_kubeconfigs")
	}

	if p.WorkspacePath == "" {
		return fmt.Errorf("workspacePath is not set")
	}

	_, err := os.Stat(p.WorkspacePath)
	if err != nil {
		return fmt.Errorf("unable to stat workspace path: %v", err)
	}

	if p.Vars == nil {
		p.Vars = make(map[string]string)
	}

	_, err = p.runTerraformCommand("init")
	if err != nil {
		return fmt.Errorf("unable to run terraform init: %v", err)
	}

	_, err = p.runTerraformCommand("apply", "-auto-approve")
	if err != nil {
		return fmt.Errorf("unable to run terraform apply: %v", err)
	}

	output, err := p.runTerraformCommandNoVars("show", "-json")
	if err != nil {
		return fmt.Errorf("unable to run terraform show: %v", err)
	}

	p.tfShowJSONBytes = output

	return nil
}

func (p *TerraformProvider) Cleanup() error {
	_, err := p.runTerraformCommand("destroy", "-auto-approve")
	if err != nil {
		return err
	}

	return nil
}
