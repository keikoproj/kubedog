# Kube

The kube package provides steps implementations related to Kubernetes.

## Templated Kubernetes manifests
All the steps implementations of this package use the [`"text/template"`](https://golang.org/pkg/text/template/) library to support templated `yaml` files. Assign the data structure that contains your template's arguments to `TemplateArguments` within the `KubeContext` object of your `Test` instance:

``` go
// Data structure with your template's arguments
type templateArgs struct {
	namespace string
}

func NewTemplateArgs() templateArgs {
	return templateArgs{
		namespace: os.Getenv("NAMESPACE"),
	}
}
```
``` go
var t kubedog.Test

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	t.SetTestSuite(ctx)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	t.SetScenario(ctx)
    // Assign your data structure
	t.KubeContext.TemplateArguments = NewTemplateArgs()
	t.Run()
}
```