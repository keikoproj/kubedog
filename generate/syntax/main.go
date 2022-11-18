package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	newLine                  = "\n"
	sourceFilePath           = "../../kubedog.go"
	actionIndicator          = "//syntax-generation"
	actionDelimiter          = ":"
	actionBegin              = "begin"
	actionEnd                = "end"
	actionTittle             = "tittle"
	tittleBeginning          = "## "
	processedStepBeginning   = "- "
	stepIndicator            = "kdt.scenarioContext.Step"
	stepDelimiter            = "`"
	stepPrefix               = "^"
	stepSuffix               = "$"
	destinationFilePath      = "../../docs/syntax.md"
	destinationFileBeginning = "# Syntax" + newLine + "Below you will find the step syntax next to the name of the method it utilizes. Here GK stands for [Gherkin](https://cucumber.io/docs/gherkin/reference/#keywords) Keyword and words in brackets ([]) are optional:" + newLine
)

var replacers = map[string]string{
	`(?:`:     `[`,
	` )?`:     `] `,
	`(\d+)`:   `<digits>`,
	`(\S+)`:   `<non-whitespace-characters>`,
	`([^"]*)`: `<any-characters-except-(")>`,
}

func main() {
	sourceFil, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(err)
	}
	defer sourceFil.Close()
	fileScanner := bufio.NewScanner(sourceFil)
	fileScanner.Split(bufio.ScanLines)
	rawSyntax := []string{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if isBeginAction(line) {
			log.Infof("found begin action as '%s'", line)
			for fileScanner.Scan() {
				line := fileScanner.Text()
				if isEndAction(line) {
					log.Infof("found end action as '%s'", line)
					break
				}
				rawSyntax = append(rawSyntax, line)
			}
			break
		}
	}
	log.Infof("found raw syntax to process as:")
	printStringSlice(rawSyntax)
	processedSyntac := processSyntax(rawSyntax)
	generateSyntax(processedSyntac)
}

func generateSyntax(processedSyntax []string) {
	if err := os.Remove(destinationFilePath); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(destinationFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Infof("writing to '%s':", destinationFilePath)
	if _, err := f.WriteString(destinationFileBeginning); err != nil {
		log.Fatal(err)
	}
	fmt.Print(destinationFileBeginning)
	for _, processedLine := range processedSyntax {
		if _, err := f.WriteString(processedLine); err != nil {
			log.Fatal(err)
		}
		fmt.Print(processedLine)
	}
}

func processSyntax(rawSyntax []string) []string {
	processedSyntax := []string{}
	for _, rawLine := range rawSyntax {
		switch {
		case strings.Contains(rawLine, actionIndicator):
			tittle := mustGetTittle(rawLine)
			processedTittle := newLine + tittleBeginning + tittle + newLine
			processedSyntax = append(processedSyntax, processedTittle)
		case strings.Contains(rawLine, stepIndicator):
			processedStep := processedStepBeginning + processStep(rawLine) + newLine
			processedSyntax = append(processedSyntax, processedStep)
		}
	}
	return processedSyntax
}

func processStep(rawStep string) string {
	if !strings.Contains(rawStep, stepIndicator) {
		log.Fatalf("expected '%s' to contain '%s'", rawStep, stepIndicator)
	}
	rawStepSplit := strings.Split(rawStep, stepDelimiter)
	if len(rawStepSplit) != 3 {
		log.Fatalf("expected '%s' to meet format '%s(%s<regexp>%s, <method>)'", rawStep, stepIndicator, stepDelimiter, stepDelimiter)
	}
	processedStep := rawStepSplit[1]
	processedStep = strings.TrimPrefix(processedStep, stepPrefix)
	processedStep = strings.TrimSuffix(processedStep, stepSuffix)
	for replacee, replacer := range replacers {
		processedStep = strings.ReplaceAll(processedStep, replacee, replacer)
	}
	return processedStep
}

func mustGetTittle(line string) string {
	action, afterAction := getAction(line)
	if action != actionTittle {
		log.Fatalf("expected '%s' to contain '%s%s%s'", line, actionIndicator, actionDelimiter, actionTittle)
	}
	if afterAction == "" {
		log.Fatalf("expected '%s' to contain '%s%s%s%s<tittle>'", line, actionIndicator, actionDelimiter, actionTittle, actionDelimiter)
	}
	return afterAction
}

func isEndAction(line string) bool {
	return isAction(actionEnd, line)
}

func isBeginAction(line string) bool {
	return isAction(actionBegin, line)
}

func isAction(expectedAction, line string) bool {
	if expectedAction == "" {
		log.Fatal("expectedAction cant be empty")
	}
	action, _ := getAction(line)
	return action == expectedAction
}

func getAction(line string) (string, string) {
	// if line == "" {
	// 	log.Fatal("line cant be empty")
	// }
	if strings.Contains(line, actionIndicator) {
		lineSplit := strings.Split(line, actionDelimiter)
		if len(lineSplit) < 2 {
			log.Fatalf("expected '%s' to contain at least 2 elements, the actionIndicator '%s' and an action separated by '%s' but got '%v'", line, actionIndicator, actionDelimiter, lineSplit)
		}
		action := lineSplit[1]
		if len(lineSplit) > 2 {
			afterAction := lineSplit[2]
			if len(lineSplit) > 3 {
				log.Warnf("'%s' had more than 3 elements delimited by '%s'. took action '%s' and afterAction '%s' and ignored the rest", line, actionDelimiter, action, afterAction)
			}
			return action, afterAction
		}
		return action, ""
	}
	return "", ""
}

func printStringSlice(slice []string) {
	for _, s := range slice {
		fmt.Print(s)
	}
}

/*
	//syntax-generation:beginning
	//syntax-generation:tittle:Generic steps
	kdt.scenarioContext.Step(`^(?:I )?wait for (\d+) (minutes|seconds)$`, common.WaitFor)
	//syntax-generation:tittle:Kubernetes steps
	kdt.scenarioContext.Step(`^((?:a )?Kubernetes cluster|(?:there are )?(?:valid )?Kubernetes Credentials)$`, kdt.KubeContext.KubernetesCluster)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+)$`, kdt.KubeContext.ResourceOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeContext.ResourceOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+)$`, kdt.KubeContext.MultiResourceOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeContext.MultiResourceOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|update) (?:the )?secret (\S+) in namespace (\S+) from (?:environment variable )?(\S+)$`, kdt.KubeContext.SecretOperationFromEnvironmentVariable)
	kdt.scenarioContext.Step(`^(?:I )?delete (?:the )?secret (\S+) in namespace (\S+)$`, kdt.KubeContext.SecretDelete)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) should be (created|deleted)$`, kdt.KubeContext.ResourceShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?Kubernetes cluster should be (created|deleted|upgraded)$`, kdt.KubeContext.KubernetesClusterShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) (?:should )?converge to selector ([^"]*)$`, kdt.KubeContext.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) condition ([^"]*) should be (true|false)$`, kdt.KubeContext.ResourceConditionShouldBe)
	kdt.scenarioContext.Step(`^(?:I )?update (?:the )?resource ([^"]*) with ([^"]*) set to ([^"]*)$`, kdt.KubeContext.UpdateResourceWithField)
	kdt.scenarioContext.Step(`^(\d+) node\(s\) with selector ([^"]*) should be (found|ready)$`, kdt.KubeContext.NodesWithSelectorShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?(deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget|sa|serviceaccount) ([^"]*) is in namespace ([^"]*)$`, kdt.KubeContext.ResourceInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?scale (?:the )?deployment ([^"]*) in namespace ([^"]*) to (\d+)$`, kdt.KubeContext.ScaleDeployment)
	kdt.scenarioContext.Step(`^(?:the )?(clusterrole|clusterrolebinding) with name ([^"]*) should be found`, kdt.KubeContext.ClusterRbacIsFound)
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?nodes list$`, kdt.KubeContext.GetNodes)
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*) with selector ([^"]*)$`, kdt.KubeContext.GetPodsWithSelector)
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*)$`, kdt.KubeContext.GetPods)
	kdt.scenarioContext.Step(`^(?:the )?(daemonset|deployment) ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeContext.ResourceIsRunning)
	kdt.scenarioContext.Step(`^(?:the )?persistentvolume ([^"]*) exists with status (Available|Bound|Released|Failed|Pending)$`, kdt.KubeContext.PersistentVolExists)
	kdt.scenarioContext.Step(`^(?:the )?(clusterrole|clusterrolebinding) with name ([^"]*) should be found$`, kdt.KubeContext.ClusterRbacIsFound)
	kdt.scenarioContext.Step(`^(?:the )?ingress (\S+) in (?:the )?namespace (\S+) (?:is )?(?:available )?on port (\d+) and path ([^"]*)$`, kdt.KubeContext.IngressAvailable)
	kdt.scenarioContext.Step(`^(?:I )?send (\d+) tps to ingress (\S+) in (?:the )?namespace (\S+) (?:available )?on port (\d+) and path ([^"]*) for (\d+) (minutes|seconds) expecting (\d+) errors$`, kdt.KubeContext.SendTrafficToIngress)
	//syntax-generation:tittle:AWS steps
	kdt.scenarioContext.Step(`^(?:there are )?(?:valid )?AWS Credentials$`, kdt.AwsContext.GetAWSCredsAndClients)
	kdt.scenarioContext.Step(`^an Auto Scaling Group named ([^"]*)$`, kdt.AwsContext.AnASGNamed)
	kdt.scenarioContext.Step(`^(?:I )?update (?:the )?current Auto Scaling Group with ([^"]*) set to ([^"]*)$`, kdt.AwsContext.UpdateFieldOfCurrentASG)
	kdt.scenarioContext.Step(`^(?:the )?current Auto Scaling Group (?:is )?scaled to \(min, max\) = \((\d+), (\d+)\)$`, kdt.AwsContext.ScaleCurrentASG)
	kdt.scenarioContext.Step(`^(?:the )?DNS name (\S+) (should|should not) be created in hostedZoneID (\S+)$`, kdt.AwsContext.DnsNameShouldOrNotInHostedZoneID)
	kdt.scenarioContext.Step(`^(?:I )?(add|remove) (?:the )?(\S+) role as trusted entity to iam role ([^"]*)$`, kdt.AwsContext.IamRoleTrust)
	kdt.scenarioContext.Step(`^(?:I )?(add|remove) ?([^"]*) as trusted entity to iam role ([^"]*)$`, kdt.AwsContext.IamRoleTrust)
	kdt.scenarioContext.Step(`^(?:I )?(add|remove) cluster shared iam role$`, kdt.AwsContext.ClusterSharedIamOperation)
	//syntax-generation:end
*/
