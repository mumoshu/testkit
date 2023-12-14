package testkit

import "testing"

// ECRImageRepositoryProvider is a provider that can provision an ECR image repository.
// Any provider that can create an ECR image repository should implement this interface.
type ECRImageRepositoryProvider interface {
	// GetECRImageRepository returns an ECR image repository.
	GetECRImageRepository(opts ...ECRImageRepositoryOption) (*ECRImageRepository, error)
}

// ECRImageRepository is an ECR image repository that is
// provided by the TestKit.
// It doesn't necessarily be created by the TestKit,
// but it can be created by other tools,
// such as eksctl.
type ECRImageRepository struct {
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

// ECRImageRepositoryOptions is the options for creating an ECR image repository.
// The zero value is a valid value.
type ECRImageRepositoryConfig struct {
	// The ID is the local name of the repository in the test,
	// which is used by the provider implmentation to identify
	// the repository in the test.
	// The ID is usually not the same as the name of the repository.
	ID string
}

type ECRImageRepositoryOption func(*ECRImageRepositoryConfig)

// ECRImageRepository creates an ECR image repository.
func (tk *TestKit) ECRImageRepository(t *testing.T, opts ...ECRImageRepositoryOption) *ECRImageRepository {
	t.Helper()

	var cp ECRImageRepositoryProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(ECRImageRepositoryProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no ECRImageRepositoryProvider found")
	}

	ecrImageRepository, err := cp.GetECRImageRepository(opts...)
	if err != nil {
		t.Fatalf("unable to get ecr image repository: %v", err)
	}

	return ecrImageRepository
}
