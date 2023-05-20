# Examples
All examples are rooted in a directory that holds a standalone [golang module](https://go.dev/ref/mod#introduction) with a [replace directive](https://go.dev/ref/mod#go-mod-file-replace) to consume the local version of `kubedog`.

1. Clone the repository
2. Move to the directory of the example you want to run
   - [`cd examples/usage`](usage)
   - [`cd examples/templating/kube`](templating/kube)
   - [`cd examples/templating/generic`](templating/generic)
3. Run 
   - `go test` for [usage](#usage) and [templating/kube](#templatingkube)
   - `go run main.go` for [templating/generic](#templatinggeneric)

To run the examples outside of the repository, you need to remove the replace directive and use a valid `kubedog` version.

## [usage](usage)

1. [usage/features/deploy-pod.feature](usage/features/deploy-pod.feature): is the `*.feature` file defining the behavior to test
2. [usage/templates](usage/templates): holds the Kubernetes yaml files, they could be templated or not
   - [namespace.yaml](usage/templates/namespace.yaml)
   - [pod.yaml](usage/templates/pod.yaml)
3. [usage/main_test.go](usage/main_test.go): is the test implementation with the minimum recommended setup for `godog` and `kubedog`

## [templating/kube](templating/kube)

1. [templating/kube/features/deploy-pod.feature](templating/kube/features/deploy-pod.feature): same as [usage](#usage)
2. [templating/kube/templates](templating/kube/templates): similar to [usage](#usage), but the files are **templated**
3. [templating/kube/main_test.go](templating/kube/main_test.go): similar to [usage](#usage), but it adds templating implementation

## [templating/generic](templating/generic)

1. [templating/generic/files](templating/generic/files): templated files
2. [templating/generic/main.go](templating/generic/main.go): templating implementation

