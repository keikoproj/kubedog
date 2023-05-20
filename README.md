# Kubedog

![Test Status](https://github.com/keikoproj/kubedog/workflows/Test/badge.svg) [![codecov](https://codecov.io/gh/keikoproj/kubedog/branch/master/graph/badge.svg)](https://codecov.io/gh/keikoproj/kubedog)

This is a simple wrapper of [Godog]( https://github.com/cucumber/godog) with some predefined steps and their implementations. It targets the [functional testing](https://cucumber.io/docs/bdd/) of [Kubernetes](https://kubernetes.io/) components working in [AWS](https://aws.amazon.com/). 

The library has a one and only purpose – to save you from the hassle of implementing steps and hooks with the compromise of using predefined syntax. But of course, you could always redefine the steps using custom syntax and pass the corresponding kubedog methods.

## Resources
- [Syntax](docs/syntax.md)
- [Examples](examples/examples.md)
- [GoDocs](https://godoc.org/github.com/keikoproj/kubedog)