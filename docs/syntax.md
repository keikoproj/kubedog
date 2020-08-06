# Syntax
Below you will find the step syntax next to the name of the method it utilizes. The implementation of the methods can be found in the [`kube`](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes) and [`aws`](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws) packages respectively. Here GK stands for [Gherkin](https://cucumber.io/docs/gherkin/reference/#keywords) keyword and words in brackets ([]) are optional:

**Kubernetes related steps**:
1. 	`<GK> a Kubernetes cluster` AKubernetesCluster
2.	`<GK> I <operation> the resource <filename>.yaml` ResourceOperation
3.	`<GK> the resource <filename> should be <state>` ResourceShouldBe
4.	`<GK> the resource <filename> [should] converge to selector <complete key>=<value>` ResourceShouldConvergeToSelector
5.	`<GK> the resource <filename> condition <condition type> should be (true|false)` ResourceConditionShouldBe
6.	`<GK> I update a resource <filename> with <complete key> set to <value>` UpdateResourceWithField
7.	`<GK>  <number of> nodes with selector <complete key>=<value> should be (found|ready)` NodesWithSelectorShouldBe

**AWS related steps**:
1.	`<GK> valid AWS Credentials` GetAWSCredsAndClients
2.	`<GK> an Auto Scaling Group named <name>` AnASGNamed
3.	`<GK> I update the current Auto Scaling Group with <field> set to <value>` UpdateFieldOfCurrentASG
4.	`<GK> the current Auto Scaling Group is scaled to (min, max) = (<min size>, <max size>)` ScaleCurrentASG