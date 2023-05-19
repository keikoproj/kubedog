# Syntax
Below you will find the step syntax next to the name of the method it utilizes. Here GK stands for [Gherkin](https://cucumber.io/docs/gherkin/reference/#keywords) Keyword and words in brackets ([]) are optional:

## Generic steps
- `<GK> [I] wait [for] <digits> (minutes|seconds)` generic.WaitFor
- `<GK> the <non-whitespace-characters> command is available` generic.CommandExists
- `<GK> I run the <non-whitespace-characters> command with the <any-characters-except-(")> args and the command (fails|succeeds)` generic.RunCommand

## Kubernetes steps
- `<GK> ([a] Kubernetes cluster|[there are] [valid] Kubernetes Credentials)` kdt.KubeClientSet.DiscoverClients
- `<GK> [the] Kubernetes cluster should be (created|deleted|upgraded)` kdt.KubeClientSet.KubernetesClusterShouldBe
- `<GK> [I] store [the] current time as <any-characters-except-(")>` kdt.KubeClientSet.SetTimestamp

### Unstructured Resources
- `<GK> [I] (create|submit|delete|update) [the] resource <non-whitespace-characters>` kdt.KubeClientSet.ResourceOperation
- `<GK> [I] (create|submit|delete|update) [the] resource <non-whitespace-characters> in [the] <any-characters-except-(")> namespace` kdt.KubeClientSet.ResourceOperationInNamespace
- `<GK> [I] (create|submit|delete|update) [the] resources in <non-whitespace-characters>` kdt.KubeClientSet.ResourcesOperation
- `<GK> [I] (create|submit|delete|update) [the] resources in <non-whitespace-characters> in [the] <any-characters-except-(")> namespace` kdt.KubeClientSet.ResourcesOperationInNamespace
- `<GK> [I] (create|submit|delete|update) [the] resource <non-whitespace-characters>, the operation should (succeed|fail)` kdt.KubeClientSet.ResourceOperationWithResult
- `<GK> [I] (create|submit|delete|update) [the] resource <non-whitespace-characters> in [the] <any-characters-except-(")> namespace, the operation should (succeed|fail)` kdt.KubeClientSet.ResourceOperationWithResultInNamespace
- `<GK> [the] resource <any-characters-except-(")> should be (created|deleted)` kdt.KubeClientSet.ResourceShouldBe
- `<GK> [the] resource <any-characters-except-(")> [should] converge to selector <non-whitespace-characters>` kdt.KubeClientSet.ResourceShouldConvergeToSelector
- `<GK> [the] resource <any-characters-except-(")> condition <any-characters-except-(")> should be (true|false)` kdt.KubeClientSet.ResourceConditionShouldBe
- `<GK> [I] update [the] resource <any-characters-except-(")> with <any-characters-except-(")> set to <any-characters-except-(")>` kdt.KubeClientSet.UpdateResourceWithField
- `<GK> [I] verify InstanceGroups [are] in "ready" state` kdt.KubeClientSet.VerifyInstanceGroups

### Structured Resources

#### Pods
- `<GK> [I] get [the] pods in namespace <any-characters-except-(")>` kdt.KubeClientSet.Pods
- `<GK> [I] get [the] pods in namespace <any-characters-except-(")> with selector <non-whitespace-characters>` kdt.KubeClientSet.PodsWithSelector
- `<GK> [the] pods in namespace <any-characters-except-(")> with selector <non-whitespace-characters> have restart count less than <digits>` kdt.KubeClientSet.PodsWithSelectorHaveRestartCountLessThan
- `<GK> (some|all) pods in namespace <non-whitespace-characters> with selector <non-whitespace-characters> have "<any-characters-except-(")>" in logs since <any-characters-except-(")> time` kdt.KubeClientSet.SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime
- `<GK> some pods in namespace <non-whitespace-characters> with selector <non-whitespace-characters> don't have "<any-characters-except-(")>" in logs since <any-characters-except-(")> time` kdt.KubeClientSet.SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime
- `<GK> [the] pods in namespace <non-whitespace-characters> with selector <non-whitespace-characters> have no errors in logs since <any-characters-except-(")> time` kdt.KubeClientSet.PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime
- `<GK> [the] pods in namespace <non-whitespace-characters> with selector <non-whitespace-characters> have some errors in logs since <any-characters-except-(")> time` kdt.KubeClientSet.PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime
- `<GK> [the] pods in namespace <non-whitespace-characters> with selector <non-whitespace-characters> should have labels <non-whitespace-characters>` kdt.KubeClientSet.PodsInNamespaceWithSelectorShouldHaveLabels
- `<GK> [the] pod <non-whitespace-characters> in namespace <non-whitespace-characters> should have labels <non-whitespace-characters>` kdt.KubeClientSet.PodInNamespaceShouldHaveLabels

#### Others
- `<GK> [I] (create|submit|update) [the] secret <non-whitespace-characters> in namespace <non-whitespace-characters> from [environment variable] <non-whitespace-characters>` kdt.KubeClientSet.SecretOperationFromEnvironmentVariable
- `<GK> [I] delete [the] secret <non-whitespace-characters> in namespace <non-whitespace-characters>` kdt.KubeClientSet.SecretDelete
- `<GK> <digits> node[s] with selector <non-whitespace-characters> should be (found|ready)` kdt.KubeClientSet.NodesWithSelectorShouldBe
- `<GK> [the] (deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget|sa|serviceaccount) <any-characters-except-(")> is in namespace <any-characters-except-(")>` kdt.KubeClientSet.ResourceInNamespace
- `<GK> [I] scale [the] deployment <any-characters-except-(")> in namespace <any-characters-except-(")> to <digits>` kdt.KubeClientSet.ScaleDeployment
- `<GK> [I] validate Prometheus Statefulset <any-characters-except-(")> in namespace <any-characters-except-(")> has volumeClaimTemplates name <any-characters-except-(")>` kdt.KubeClientSet.ValidatePrometheusVolumeClaimTemplatesName
- `<GK> [I] get [the] nodes list` kdt.KubeClientSet.GetNodes
- `<GK> [the] daemonset <any-characters-except-(")> is running in namespace <any-characters-except-(")>` kdt.KubeClientSet.DaemonSetIsRunning
- `<GK> [the] deployment <any-characters-except-(")> is running in namespace <any-characters-except-(")>` kdt.KubeClientSet.DeploymentIsRunning
- `<GK> [the] persistentvolume <any-characters-except-(")> exists with status (Available|Bound|Released|Failed|Pending)` kdt.KubeClientSet.PersistentVolExists
- `<GK> [the] (clusterrole|clusterrolebinding) with name <any-characters-except-(")> should be found` kdt.KubeClientSet.ClusterRbacIsFound
- `<GK> [the] ingress <non-whitespace-characters> in [the] namespace <non-whitespace-characters> [is] [available] on port <digits> and path <any-characters-except-(")>` kdt.KubeClientSet.IngressAvailable
- `<GK> [I] send <digits> tps to ingress <non-whitespace-characters> in [the] namespace <non-whitespace-characters> [available] on port <digits> and path <any-characters-except-(")> for <digits> (minutes|seconds) expecting up to <digits> error[s]` kdt.KubeClientSet.SendTrafficToIngress

## AWS steps
- `<GK> [there are] [valid] AWS Credentials` kdt.AwsClientSet.DiscoverClients
- `<GK> an Auto Scaling Group named <any-characters-except-(")>` kdt.AwsClientSet.AnASGNamed
- `<GK> [I] update [the] current Auto Scaling Group with <any-characters-except-(")> set to <any-characters-except-(")>` kdt.AwsClientSet.UpdateFieldOfCurrentASG
- `<GK> [the] current Auto Scaling Group [is] scaled to (min, max) = (<digits>, <digits>)` kdt.AwsClientSet.ScaleCurrentASG
- `<GK> [the] DNS name <non-whitespace-characters> (should|should not) be created in hostedZoneID <non-whitespace-characters>` kdt.AwsClientSet.DnsNameShouldOrNotInHostedZoneID
- `<GK> [I] (add|remove) [the] <non-whitespace-characters> role as trusted entity to iam role <any-characters-except-(")>` kdt.AwsClientSet.IamRoleTrust
- `<GK> [I] (add|remove) cluster shared iam role` kdt.AwsClientSet.ClusterSharedIamOperation
