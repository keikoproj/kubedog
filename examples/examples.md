# Example:

Let's use [Upgrade Manager functional test](https://github.com/keikoproj/upgrade-manager/tree/master/test-bdd) to easily show how to setup Kubedog around Godog. The example below assumes you know how Godog works. Some idea of what [Upgrade Manager](https://github.com/keikoproj/upgrade-manager) is, would help, but it is not necessary. 

Let’s jump right into it:

## Step 1: the feature file

First we need the `*.feature` file defining the desired behavior. You can define this file as you would normally do when using Godog, but utilizing [Kubedog syntax](syntax.md). As mentioned, you can also redefine the steps and use custom syntax. 

``` gherkin
Feature: UM's RollingUpgrade Create
  In order to create RollingUpgrades
  As an EKS cluster operator
  I need to submit the custom resource

  Background:
    Given valid AWS Credentials
    And a Kubernetes cluster
    And an Auto Scaling Group named upgrademgr-eks-nightly-ASG

  Scenario: The ASG had a launch config update that allows nodes to join
    Given the current Auto Scaling Group has the required initial settings
    Then 1 node(s) with selector bdd-test=preUpgrade-label should be ready
    Given I update the current Auto Scaling Group with LaunchConfigurationName set to upgrade-eks-nightly-LC-postUpgrade
    And I submit the resource rolling-upgrade.yaml
    Then the resource rolling-upgrade.yaml should be created
    When the resource rolling-upgrade.yaml converge to selector .status.currentStatus=completed
    Then 1 node(s) with selector bdd-test=postUpgrade-label should be ready
```
## Step 2: setting Kubedog around Godog

### Minimum required setup

We would need the functions `InitializeTestSuite` and `InitializeScenario` in the `*_test.go` file as required by Godog. We have to pass the suite and scenario context pointers with the methods `SetTestSuite` and `SetScenario` of `kubedog.Test` and call the `Run` method:

``` go
var t kubedog.Test

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	t.SetTestSuite(ctx)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	t.SetScenario(ctx)
	t.Run()
}
```

### Before and After Suite, Scenario and Step hooks

Again, kubedog is a simple wrapper – if you want to take advantage of the hooks, you can do so by defining your own functions or calling kubedog’s functions/methods.

```
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
```
``` go
	ctx.BeforeSuite(func() {
		log.Info("BDD >> trying to delete any existing test RollingUpgrade")
		err := t.KubeContext.DeleteAllTestResources()
		if err != nil {
			log.Errorf("Failed deleting the test resources: %v", err)
		}
	})

	ctx.AfterSuite(func() {
		log.Infof("BDD >> scaling down the ASG %v", t.AwsContext.AsgName)
		err := t.AwsContext.ScaleCurrentASG(0, 0)
		if err != nil {
			log.Errorf("Failed scaling down the ASG %v: %v", t.AwsContext.AsgName, err)
		}

		log.Info("BDD >> deleting any existing test RollingUpgrade")
		err = t.KubeContext.DeleteAllTestResources()
		if err != nil {
			log.Errorf("Failed deleting the test resources: %v", err)
		}
	})
```
```
	t.SetTestSuite(ctx)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
```
``` go
	ctx.AfterStep(func(s *godog.Step, err error) {
		time.Sleep(time.Second * 5)
	})
```
```
	t.SetScenario(ctx)
	t.Run()
}
```

### Custom steps definition

You are welcome to define new steps and pass kubedog's methods or define your own functions:

```
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.AfterStep(func(s *godog.Step, err error) {
		time.Sleep(time.Second * 5)
	})
```
``` go
	ctx.Step(`^the current Auto Scaling Group has the required initial settings$`, theRequiredInitialSettings)
```
```
	t.SetScenario(ctx)
	t.Run()
}
```
``` go
func theRequiredInitialSettings() error {
	// Making sure the ASG has the pre-test launch config and 1 node with correct config
	err := t.AwsContext.UpdateFieldOfCurrentASG("LaunchConfigurationName", "upgrade-eks-nightly-LC-preUpgrade")
	if err != nil {
		return err
	}
	err = t.AwsContext.ScaleCurrentASG(0, 0)
	if err != nil {
		return err
	}
	err = t.KubeContext.NodesWithSelectorShouldBe(0, "bdd-test=preUpgrade-label", "found")
	if err != nil {
		return err
	}
	err = t.KubeContext.NodesWithSelectorShouldBe(0, "bdd-test=postUpgrade-label", "found")
	if err != nil {
		return err
	}
	err = t.AwsContext.ScaleCurrentASG(1, 1)
	if err != nil {
		return err
	}
	return nil
}
```

### Golang test

We can also set up [`go test` compatibility](https://github.com/keikoproj/upgrade-manager/blob/master/test-bdd/main_test.go#L15), as explained in [Godog’s repository](https://github.com/cucumber/godog#running-godog-with-go-test).