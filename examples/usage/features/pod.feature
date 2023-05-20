Feature: Successfully deploying a Kubernetes Pod

  Background: Valid Credentials
    Given valid AWS Credentials
    And a Kubernetes cluster

  Scenario: Create Namespace and Pod, validate successfull deployment
    Then create resource generated_namespace.yaml
    And store current time as pod-creation
    And create resource generated_pod.yaml
    Then resource generated_pod.yaml should be created
    And resource generated_pod.yaml condition Ready should be True
    Then all pods in namespace kubedog-example with selector tier=backend have "Hello, Kubedog!" in logs since pod-creation time