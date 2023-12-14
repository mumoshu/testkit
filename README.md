# mumoshu/testkit

`mumoshu/testkit` is a toolkit for writing Go tests for your cloud applications.

Currently, this provides the following tools:

- Abstraction over terraform/eksctl/envvars/kubectl/k8s-kind so that you can provision/retain/destroy the cloud test harness for max developer productivity
- Various helpers for writing test assertions against cloud resources

It is handy when you want to write an integration or E2E tests for:

- Kubernetes controllers
- Terraform workspaces
- ChatOps bots
- Complex CI/CD workflows
- AWS CDK projects (Planned)
- Pulumi projects (Planned)

See [testkit_test.go](testkit_test.go) for inspiration on how you would write tests with `testkit`.
