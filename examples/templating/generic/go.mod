module github.com/keikoproj/kubedog/examples/templating/generic

go 1.21

replace github.com/keikoproj/kubedog => ../../../

require github.com/keikoproj/kubedog v1.2.3

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
	k8s.io/apimachinery v0.28.12 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
)
