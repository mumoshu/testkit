// mumoshu/testkit is a set of tools for testing
// your application and/or infrastructure code end-to-end.
//
// For application testing, it can provision a real infrastructure
// in the cloud, runs your code on it, and verifies the results.
// It supports AWS EKS and S3, GitHub, Slack.
//
// For infrastructure testing, it calls Terraform, eksctl, etc., of
// your choice to provision the infrastructure, optionally deploy test agents on it,
// and finally runs your tests against it.
//
// It is designed to abstract away and automate the common patterns
// of testing infrastructure code, so that you can focus on writing
// the actual tests.
//
// For example, it can delegate the creation of an EKS cluster to eksctl,
// and other resources to Terraform, and then run your tests against them.
// It can also create a GitHub repository, push your code to it,
// and then run your tests against it.
package testkit

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type TestKit struct {
	Config

	availableProviders []Provider
}

type Config struct {
	Providers                []Provider
	RetainResources          bool
	RetainResourcesOnFailure bool
}

type Option func(*Config)

func RetainResources() Option {
	return func(tk *Config) {
		tk.RetainResources = true
	}
}

func RetainResourcesOnFailure() Option {
	return func(tk *Config) {
		tk.RetainResourcesOnFailure = true
	}
}

func Providers(providers ...Provider) Option {
	return func(tk *Config) {
		tk.Providers = append(tk.Providers, providers...)
	}
}

// New creates a new TestKit harness.
// It fails the test if it cannot create the TestKit.
// It automatically cleans up all the resources created by the TestKit
// at the end of the test.
//
// If the RetainResources option is true, it does not clean up the resources.
// This is useful for debugging.
// If the RetainResourcesOnFailure option is true, it does not clean up the resources
// if the test fails.
// This is useful for debugging.
//
// IF the Providers option is not empty, it uses the providers specified in the option.
//
// If providers is empty, it uses the default providers.
// The default providers are the providers that are available
// in the current environment.
// Availability of a provider is determined by the Setup method.
// If no provider is available, it fails the test.
func New(t *testing.T, opts ...Option) *TestKit {
	t.Helper()

	var conf Config

	for _, opt := range opts {
		opt(&conf)
	}

	if len(conf.Providers) == 0 {
		var defaultProviders []Provider
		if len(conf.Providers) > 0 {
			defaultProviders = conf.Providers
		} else {
			defaultProviders = []Provider{
				// TODO: Omit this when cluster.yaml does not exist
				&EKSCTLProvider{},
				// TODO: Omit this when the terraform workspace is not specified
				&TerraformProvider{},
				&EnvProvider{},
			}
		}

		var providers []Provider

		for _, p := range defaultProviders {
			if err := p.Setup(); err != nil {
				t.Logf("skipped setting up failed provider %v: %v", p, err)
				continue
			}
			providers = append(providers, p)
		}

		if len(providers) == 0 {
			t.Fatal("no provider out of the default providers is available")
		}

		conf.Providers = providers
	} else {
		for _, p := range conf.Providers {
			if err := p.Setup(); err != nil {
				t.Fatalf("failed to setup provider %v: %v", p, err)
			}
		}
	}

	tk := &TestKit{
		Config:             conf,
		availableProviders: conf.Providers,
	}

	t.Cleanup(func() {
		tk.Cleanup(t)
	})

	return tk
}

// Cleanup cleans up all the resources created by the TestKit.
// The caller should call this function at the end of the test,
// typically in a defer statement.
// If the TestKit is created with the RetainResources option,
// this function does nothing.
func (tk *TestKit) Cleanup(t *testing.T) {
	if !tk.CleanupNeeded(t) {
		return
	}

	for _, p := range tk.availableProviders {
		if err := p.Cleanup(); err != nil {
			t.Logf("failed to cleanup provider %v: %v", p, err)
		}
	}
}

// CleanupNeeded returns true if the test harness needs to be cleaned up.
//
// This takes into account the RetainResources and RetainResourcesOnFailure options and the test result.
// If the RetainResources option is true, it returns false.
// If the RetainResourcesOnFailure option is true and the test failed, it returns false.
// Otherwise, it returns true.
//
// This is useful when you want to clean up resources unmanaged by the TestKit
// respecting the RetainResources and RetainResourcesOnFailure options.
func (tk *TestKit) CleanupNeeded(t *testing.T) bool {
	retainResources := tk.RetainResources || (t.Failed() && tk.RetainResourcesOnFailure)

	return !retainResources
}

type Provider interface {
	Setup() error
	Cleanup() error
}

func (s *S3Bucket) AWSV2Config(t *testing.T) aws.Config {
	t.Helper()
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		t.Fatalf("failed to load AWS SDK config: %v", err)
	}
	sdkConfig.Region = s.Region

	return sdkConfig
}
