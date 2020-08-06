#Syntax

Here GK stands for Gherkin keyword and words in brackets ([]) are optional:

**Kubernetes related steps**: methods can be found in [kube](https://godoc.org/github.com/keikoproj/kubedog/pkg/kubernetes) package
1. 	`<GK> a Kubernetes cluster`
	Method: AKubernetesCluster
2.	`<GK> I <operation> the resource <filename>.yaml`
	Method: ResourceOperation
3.	`<GK> the resource <filename> should be <state>`
	Method: ResourceShouldBe
4.	`<GK> the resource <filename> [should] converge to selector <complete key>=<value>`
	Method: ResourceShouldConvergeToSelector
5.	`<GK> the resource <filename> condition <condition type> should be (true|false)`
	Method: ResourceConditionShouldBe
6.	`<GK> I update a resource <filename> with <complete key> set to <value>`
	Method: UpdateResourceWithField
7.	`<GK>  <number of> nodes with selector <complete key>=<value> should be (found|ready)`
	Method: NodesWithSelectorShouldBe

**AWS related steps**: methods can be found in [aws](https://godoc.org/github.com/keikoproj/kubedog/pkg/aws) package
1.	`<GK> valid AWS Credentials`
	Method: GetAWSCredsAndClients
2.	`<GK> an Auto Scaling Group named <name>`
	Method: AnASGNamed
3.	`<GK> I update the current Auto Scaling Group with <field> set to <value>`
	Method: UpdateFieldOfCurrentASG
4.	`<GK> the current Auto Scaling Group is scaled to (min, max) = (<min size>, <max size>)`
	Method: ScaleCurrentASG