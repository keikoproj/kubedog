package kube

import (
	"os"
	"path/filepath"
	"time"

	unstruct "github.com/keikoproj/kubedog/pkg/kubernetes/unstructured"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// TODO: this moved to unstructured pkg, check what is still needed or not
const (
	operationCreate = "create"
	operationSubmit = "submit"
	operationUpdate = "update"
	operationDelete = "delete"

	stateCreated  = "created"
	stateDeleted  = "deleted"
	stateUpgraded = "upgraded"
	stateReady    = "ready"
	stateFound    = "found"
)

type ClientSet struct {
	KubeInterface      kubernetes.Interface
	DynamicInterface   dynamic.Interface
	DiscoveryInterface discovery.DiscoveryInterface
	FilesPath          string
	TemplateArguments  interface{}
	WaiterInterval     time.Duration
	WaiterTries        int
	Timestamps         map[string]time.Time
}

func (kc *ClientSet) Validate() error {
	commonMessage := "'DiscoverClients' sets this interface, try calling it before using this method"
	if kc.DynamicInterface == nil {
		return errors.Errorf("'ClientSet.DynamicInterface' is nil. %s", commonMessage)
	}
	if kc.DiscoveryInterface == nil {
		return errors.Errorf("'ClientSet.DiscoveryInterface' is nil. %s", commonMessage)
	}
	if kc.KubeInterface == nil {
		return errors.Errorf("'ClientSet.KubeInterface' is nil. %s", commonMessage)
	}
	return nil
}

func (kc *ClientSet) DiscoverClients() error {
	var (
		home, _        = os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	)

	if exported := os.Getenv("KUBECONFIG"); exported != "" {
		kubeconfigPath = exported
	}

	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return errors.Errorf("[KUBEDOG] expected kubeconfig to exist for create operation, '%v'", kubeconfigPath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct dynamic client", err)
	}

	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	kc.KubeInterface = client
	kc.DynamicInterface = dynClient
	kc.DiscoveryInterface = discoveryClient

	return nil
}

func (kc *ClientSet) SetTimestamp(timestampName string) error {
	if kc.Timestamps == nil {
		kc.Timestamps = map[string]time.Time{}
	}
	now := time.Now()
	kc.Timestamps[timestampName] = now
	log.Infof("Memorizing '%s' time is %v", timestampName, now)
	return nil
}

func (kc *ClientSet) getResourcePath(resourceFileName string) string {
	templatesPath := kc.getTemplatesPath()
	return filepath.Join(templatesPath, resourceFileName)
}

func (kc *ClientSet) getTemplatesPath() string {
	defaultFilePath := "templates"
	if kc.FilesPath != "" {
		return kc.FilesPath
	}
	return defaultFilePath
}

func (kc *ClientSet) getWaiterInterval() time.Duration {
	defaultWaiterInterval := time.Second * 30
	if kc.WaiterInterval > 0 {
		return kc.WaiterInterval
	}
	return defaultWaiterInterval
}

func (kc *ClientSet) getWaiterTries() int {
	defaultWaiterTries := 40
	if kc.WaiterTries > 0 {
		return kc.WaiterTries
	}
	return defaultWaiterTries
}

func (kc *ClientSet) getWaiterConfig() unstruct.WaiterConfig {
	return unstruct.NewWaiterConfig(kc.getWaiterTries(), kc.getWaiterInterval())
}

func (kc *ClientSet) DeleteAllTestResources() error {
	return unstruct.DeleteResourcesAtPath(kc.DynamicInterface,
		kc.DiscoveryInterface,
		kc.TemplateArguments,
		kc.getWaiterConfig(),
		kc.getTemplatesPath())
}

func (kc *ClientSet) ResourceOperation(operation, resourceFileName string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperation(kc.DynamicInterface, resource, operation)
}

func (kc *ClientSet) ResourceOperationInNamespace(operation, resourceFileName, ns string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationInNamespace(kc.DynamicInterface, resource, operation, ns)
}

func (kc *ClientSet) MultiResourceOperation(operation, resourcesFileName string) error {
	resources, err := unstruct.GetResources(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourcesFileName))
	if err != nil {
		return err
	}
	return unstruct.MultiResourceOperation(kc.DynamicInterface, resources, operation)
}

func (kc *ClientSet) MultiResourceOperationInNamespace(operation, resourcesFileName, ns string) error {
	resources, err := unstruct.GetResources(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourcesFileName))
	if err != nil {
		return err
	}
	return unstruct.MultiResourceOperationInNamespace(kc.DynamicInterface, resources, operation, ns)
}

func (kc *ClientSet) ResourceOperationWithResult(operation, resourceFileName, expectedResult string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationWithResult(kc.DynamicInterface, resource, operation, expectedResult)
}

func (kc *ClientSet) ResourceOperationWithResultInNamespace(operation, resourceFileName, ns, expectedResult string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationWithResultInNamespace(kc.DynamicInterface, resource, operation, ns, expectedResult)
}

func (kc *ClientSet) ResourceShouldBe(resourceFileName, state string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceShouldBe(kc.DynamicInterface, resource, kc.getWaiterConfig(), state)
}

func (kc *ClientSet) ResourceShouldConvergeToSelector(resourceFileName, selector string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceShouldConvergeToSelector(kc.DynamicInterface, resource, kc.getWaiterConfig(), selector)
}

func (kc *ClientSet) ResourceConditionShouldBe(resourceFileName, conditionType, conditionValue string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceConditionShouldBe(kc.DynamicInterface, resource, kc.getWaiterConfig(), conditionType, conditionValue)
}

func (kc *ClientSet) UpdateResourceWithField(resourceFileName, key, value string) error {
	resource, err := unstruct.GetResource(kc.DiscoveryInterface, kc.TemplateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.UpdateResourceWithField(kc.DynamicInterface, resource, key, value)
}
