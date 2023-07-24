```
.
├── LICENSE
├── Makefile
├── README.md
├── coverage.txt
├── docs
│   ├── examples.md
│   └── syntax.md
├── examples
│   ├── templating
│   │   ├── generic
│   │   │   ├── files
│   │   │   │   ├── generated_pod.yaml
│   │   │   │   └── pod.yaml
│   │   │   ├── go.mod
│   │   │   ├── go.sum
│   │   │   └── main.go
│   │   └── kube
│   │       ├── features
│   │       │   └── deploy-pod.feature
│   │       ├── go.mod
│   │       ├── go.sum
│   │       ├── main_test.go
│   │       └── templates
│   │           ├── namespace.yaml
│   │           └── pod.yaml
│   └── usage
│       ├── features
│       │   └── deploy-pod.feature
│       ├── go.mod
│       ├── go.sum
│       ├── main_test.go
│       └── templates
│           ├── namespace.yaml
│           └── pod.yaml
├── generate
│   └── syntax
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   └── util
│       └── util.go
├── kubedog
├── kubedog-design.md
├── kubedog.go
└── pkg
    ├── aws
    │   ├── aws.go
    │   ├── aws_helper.go
    │   ├── aws_helper_test.go
    │   ├── aws_test.go
    │   └── iam
    │       ├── iam.go
    │       ├── iam_helper.go
    │       ├── iam_helper_test.go
    │       └── iam_test.go
    ├── generic
    │   ├── generic.go
    │   ├── generic_test.go
    │   ├── template.go
    │   ├── template_test.go
    │   └── test
    │       ├── generated_templated-bad-kind.yaml
    │       ├── generated_templated.yaml
    │       ├── templated-bad-kind.yaml
    │       └── templated.yaml
    └── kube
        ├── common
        │   └── common.go
        ├── kube.go
        ├── kube_helper.go
        ├── pod
        │   ├── pod.go
        │   ├── pod_helper.go
        │   └── pod_test.go
        ├── structured
        │   ├── structured.go
        │   ├── structured_helper.go
        │   └── structured_test.go
        └── unstructured
            ├── test
            │   ├── files
            │   │   ├── analysis-template.yaml
            │   │   ├── instance-group-not-ready.yaml
            │   │   ├── instance-group.yaml
            │   │   ├── multi-resource-no-ns.yaml
            │   │   ├── multi-resource.yaml
            │   │   ├── resource-no-ns.yaml
            │   │   └── resource.yaml
            │   └── templates
            │       ├── generated_templated.yaml
            │       └── templated.yaml
            ├── unstructured.go
            ├── unstructured_helper.go
            └── unstructured_test.go

29 directories, 67 files
```

# Kubedog's Design
<!--
rename to structure and not design?
-->

```
.
├── kubedog.go
├── pkg
│   ├── kube
│   │   ├── common
│   │   │   └── common.go
│   │   ├── kube.go
│   │   ├── kube_helper.go
│   │   ├── pod
│   │   │   ├── pod.go
│   │   │   ├── pod_helper.go
│   │   │   └── pod_test.go
│   │   ├── structured
│   │   │   ├── structured.go
│   │   │   ├── structured_helper.go
│   │   │   └── structured_test.go
│   │   └── unstructured
│   │       ├── test
│   │       │   ├── files
│   │       │   │   ├── analysis-template.yaml
│   │       │   │   ├── instance-group-not-ready.yaml
│   │       │   │   ├── instance-group.yaml
│   │       │   │   ├── multi-resource-no-ns.yaml
│   │       │   │   ├── multi-resource.yaml
│   │       │   │   ├── resource-no-ns.yaml
│   │       │   │   └── resource.yaml
│   │       │   └── templates
│   │       │       └── templated.yaml
│   │       ├── unstructured.go
│   │       ├── unstructured_helper.go
│   │       └── unstructured_test.go
│   ├── aws
│   │   ├── aws.go
│   │   ├── aws_helper.go
│   │   ├── aws_helper_test.go
│   │   ├── aws_test.go
│   │   └── iam
│   │       ├── iam.go
│   │       ├── iam_helper.go
│   │       ├── iam_helper_test.go
│   │       └── iam_test.go
│   └── generic
│       ├── generic.go
│       ├── generic_test.go
│       ├── template.go
│       ├── template_test.go
│       └── test
│           ├── templated-bad-kind.yaml
│           └── templated.yaml
└── internal
    └── util
        └── util.go
```

## Kubedog Structure

Kubedog follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

```
.
├── kubedog.go
├── pkg
└── internal
```
`kubedog.go`: godog step's definitions linking syntax with implementation.
<!--
 godog centric. 
 Syntax centric.
 Avoid code implementation here.
 Note: in the syntax context the verb 'get' lists and fails if not found. In code context the verb 'get' returns an object and 'list' lists and fails if no found.
-->
```
.
├── kubedog.go
├── pkg
│   ├── kube
│   ├── aws
│   └── generic
└── internal
    └── util
```

`kube`: Kubernetes related steps implementations
`aws`: Amazon Web Services related steps implementations
`generic`: non Kubernetes nor AWS related steps implementations
`internal/util`: utilitarian implementation to aid packages with steps implementations

### Package Structure

```
pkg/kube
├── unstructured
├── structured
├── pod
├── common
└── kube.go
```

`kube.go` imports all the sub-packages. Implementation here should be minimal and centric around Kubernetes clients, taking the clients and the behavior-changing inputs to call the sub-packages behavior specific methods accordingly.

`pod`, `structured` and `unstructured` are sub-packages with implementation grouped by subject (e.g. type: pod; data-structure category: structured or unstructured). `common` has implementation that is common among the other sub-packages.

### Sub-package Structure

```
pkg/kube/unstructured
├── unstructured.go
├── unstructured_helper.go
└── unstructured_test.go
```

`unstructured.go`: steps implementations only.
`unstructured_helper.go`: aids steps implementations and other non steps related implementations
<!-- 
TODO: how about kube_helper.go? maybe merge with kube.go or move non steps related implementations from kube.go to kube_helper.go
-->

<!--
TODO: talk about the `test` directory and the sub-directory

    ├── test
    │   ├── files
    │   │   ├── analysis-template.yaml
    │   │   ├── instance-group-not-ready.yaml
    │   │   ├── instance-group.yaml
    │   │   ├── multi-resource-no-ns.yaml
    │   │   ├── multi-resource.yaml
    │   │   ├── resource-no-ns.yaml
    │   │   └── resource.yaml
    │   └── templates
    │       ├── generated_templated.yaml
    │       └── templated.yaml

-->

### aws

```
pkg/aws
├── aws.go
├── aws_helper.go
├── aws_helper_test.go
├── aws_test.go
└── iam
    ├── iam.go
    ├── iam_helper.go
    ├── iam_helper_test.go
    └── iam_test.go
```

<!--
TODO: move *_helper_test.go to *_test.go
-->

### generic

```
pkg/generic
├── generic.go
├── generic_test.go
├── template.go
├── template_test.go
└── test
    ├── generated_templated-bad-kind.yaml
    ├── generated_templated.yaml
    ├── templated-bad-kind.yaml
    └── templated.yaml
```

<!--
TODO: move generic.go and template.go to sub-packages (directories)
-->

### util
```
internal/util
└── util.go
```