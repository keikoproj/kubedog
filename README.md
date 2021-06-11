# Kubedog

![Test Status](https://github.com/keikoproj/kubedog/workflows/Test/badge.svg) [![codecov](https://codecov.io/gh/keikoproj/kubedog/branch/master/graph/badge.svg)](https://codecov.io/gh/keikoproj/kubedog)

This is a simple wrapper of [Godog]( https://github.com/cucumber/godog) v0.10.0 with some predefined steps and their implementations. It is limited to [Kubernetes](https://kubernetes.io/) and [AWS](https://aws.amazon.com/), more precisely, it targets the [functional testing](https://cucumber.io/docs/bdd/) of Kubernetes components working in AWS. Additionally, you can take advantage of Godog’s before and after suite, scenario and step hooks and call Kubedog’s functions and methods from there. 

The library has a one and only purpose – it saves you from the hassle of implementing common/basic steps and hooks with the compromise of using predefined syntax. Of course, you could always redefine the steps using custom syntax and pass the corresponding kubedog methods.

## Resources:

- [Steps syntax](https://github.com/keikoproj/kubedog/blob/master/docs/syntax.md)
- [Usage example](https://github.com/keikoproj/kubedog/blob/master/docs/example.md): upgrade-manager's BDD

#### GoDocs

- [Kubedog](https://godoc.org/github.com/keikoproj/kubedog)
- [Kube](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes)
- [AWS](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws)