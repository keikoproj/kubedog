package generate

import (
	"github.com/keikoproj/kubedog/pkg/generic"
)

func Templates() error {
	templateArguments := []generic.TemplateArgument{
		{
			Key:                 "Namespace",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_NAMESPACE",
			Mandatory:           false,
			Default:             "kubedog-example",
		},
		{
			Key:                 "Image",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_IMAGE",
			Mandatory:           false,
			Default:             "busybox:1.28",
		},
		{
			Key:                 "Message",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_MESSAGE",
			Mandatory:           false,
			Default:             "Hello, Kubedog!",
		},
	}
	args, err := generic.TemplateArgumentsToMap(templateArguments...)
	if err != nil {
		return err
	}
	templatedPodFilePath := "templates/pod.yaml"
	_, err = generic.GenerateFileFromTemplate(templatedPodFilePath, args)
	if err != nil {
		return err
	}
	templatedNamespaceFilePath := "templates/namespace.yaml"
	_, err = generic.GenerateFileFromTemplate(templatedNamespaceFilePath, args)
	if err != nil {
		return err
	}
	return nil
}
