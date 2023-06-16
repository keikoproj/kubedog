module github.com/keikoproj/kubedog/examples/templating/generic

go 1.19

replace github.com/keikoproj/kubedog => ../../../

require github.com/keikoproj/kubedog v1.2.3

require (
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	k8s.io/apimachinery v0.22.17 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
)
