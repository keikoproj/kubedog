package main

import (
	"log"

	"github.com/keikoproj/kubedog/pkg/generic"
)

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
	templatedFilePath := "templates/pod.yaml"
	_, err = generic.GenerateFileFromTemplate(templatedFilePath, args)
	if err != nil {
		log.Fatalln(err)
	}
}
