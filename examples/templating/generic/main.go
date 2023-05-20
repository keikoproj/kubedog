package main

import (
	"log"

	"github.com/keikoproj/kubedog/pkg/generic"
)

// Using the generic package to:
// - get template arguments from environment variables
// - generate file from template using the obtained arguments
func main() {
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
		log.Fatalln(err)
	}
	templatedFilePath := "files/pod.yaml"
	_, err = generic.GenerateFileFromTemplate(templatedFilePath, args)
	if err != nil {
		log.Fatalln(err)
	}
}
