# Syntax
Below you will find the step syntax next to the name of the method it utilizes. The implementation of the methods can be found in the [`kube`](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes) and [`aws`](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws) packages respectively. Here GK stands for [Gherkin](https://cucumber.io/docs/gherkin/reference/#keywords) Keyword and words in brackets ([]) are optional:

**Kubernetes related steps**:
1. 	`<GK> a Kubernetes cluster` [AKubernetesCluster](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.AKubernetesCluster)
2.	`<GK> [I] <create|submit|delete|update> [the] resource <filename>.yaml` [ResourceOperation](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceOperation)
3.	`<GK> [I] <create|submit|delete|update> [the] resource <filename>.yaml in [the] <namespacename> namespace` [ResourceOperationInNamespace](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceOperationInNamespace)
4.	`<GK> [I] <create|submit|delete|update> [the] resources in <filename>.yaml` [MultiResourceOperation](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.MultiResourceOperation)
5.	`<GK> [I] <create|submit|delete|update> [the] resources in <filename>.yaml in [the] <namespacename> namespace` [MultiResourceOperationInNamespace](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.MultiResourceOperationInNamespace)
6.	`<GK> [the] resource <filename> should be <created|deleted>` [ResourceShouldBe](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceShouldBe)
7.	`<GK> [the] resource <filename> [should] converge(d) to selector <complete key>=<value>` [ResourceShouldConvergeToSelector](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceShouldConvergeToSelector)
8.	`<GK> [the] resource <filename> condition <condition type> should be (true|false)` [ResourceConditionShouldBe](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceConditionShouldBe)
9.	`<GK> [I] update [a] resource <filename> with <complete key> set to <value>` [UpdateResourceWithField](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.UpdateResourceWithField)
10.	`<GK>  <number of> nodes with selector <complete key>=<value> should be (found|ready)` [NodesWithSelectorShouldBe](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes#Client.NodesWithSelectorShouldBe)
11. `<GK>  [the] <deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget> <resource_name> is in namespace <namespace-name>` [ResourceInNamespace](https://pkg.go.dev/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ResourceInNamespace)
12. `<GK>  [I] scale [the] deployment <filename> in namespace <namespace-name> to <replica(s)>` [ScaleDeployment](https://pkg.go.dev/github.com/keikoproj/kubedog/pkg/kubernetes#Client.ScaleDeployment)

**AWS related steps**:
1.	`<GK> valid AWS Credentials` [GetAWSCredsAndClients](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws#Client.GetAWSCredsAndClients)
2.	`<GK> an Auto Scaling Group named <name>` [AnASGNamed](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws#Client.AnASGNamed)
3.	`<GK> [I] update [the] current Auto Scaling Group with <field> set to <value>` [UpdateFieldOfCurrentASG](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws#Client.UpdateFieldOfCurrentASG)
4.	`<GK> [the] current Auto Scaling Group [is] scaled to (min, max) = (<min size>, <max size>)` [ScaleCurrentASG](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws#Client.ScaleCurrentASG)
