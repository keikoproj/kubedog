# Examples
All examples are rooted in a directory that holds a standalone [golang module](https://go.dev/ref/mod#introduction) with a [replace directive](https://go.dev/ref/mod#go-mod-file-replace) to consume the local version of `kubedog`.

1. Clone the repository
2. Move to the directory of the example you want to run
   - [`cd examples/usage`](../examples/usage)
   - [`cd examples/templating/kube`](../examples/templating/kube)
   - [`cd examples/templating/generic`](../examples/templating/generic)
3. Run 
   - `go test` for [usage](#usage) and [templating/kube](#templatingkube)
   - `go run main.go` for [templating/generic](#templatinggeneric)

To run the examples outside of the repository, you need to remove the replace directive and use a valid `kubedog` version.

## usage

1. [usage/features/deploy-pod.feature](../examples/usage/features/deploy-pod.feature): is the `*.feature` file defining the behavior to test
2. [usage/templates](../examples/usage/templates): holds the Kubernetes yaml files, they could be templated
   - [namespace.yaml](../examples/usage/templates/namespace.yaml)
   - [pod.yaml](../examples/usage/templates/pod.yaml)
3. [usage/main_test.go](../examples/usage/main_test.go): is the test implementation with the minimum recommended setup for `godog` and `kubedog`

## templating/kube

The [kube](../pkg/kube) package has built-in templating support, this example showcases that.

1. [templating/kube/features/deploy-pod.feature](../examples/templating/kube/features/deploy-pod.feature): same as [usage](#usage)
2. [templating/kube/templates](../examples/templating/kube/templates): similar to [usage](#usage), but the files are **templated**
3. [templating/kube/main_test.go](../examples/templating/kube/main_test.go): similar to [usage](#usage), but it adds templating implementation

## templating/generic

The [generic](../pkg/generic/template.go) package offers general purpose file templating, this example showcases that.

1. [templating/generic/files](../examples/templating/generic/files): templated files
2. [templating/generic/main.go](../examples/templating/generic/main.go): templating implementation

