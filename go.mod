module github.com/keikoproj/kubedog

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.8
	github.com/cucumber/godog v0.10.0
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)

replace github.com/keikoproj/kubedog => github.com/kyle-wong/kubedog v0.1.3-0.20210728221204-b1d2ca828b0e
