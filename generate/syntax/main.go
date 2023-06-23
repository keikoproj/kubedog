/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	sourceFilePath           = "kubedog.go"
	destinationFilePath      = "docs/syntax.md"
	newLine                  = "\n"
	actionIndicator          = "//syntax-generation"
	actionDelimiter          = ":"
	actionBegin              = "begin"
	actionEnd                = "end"
	actionTitle              = "title"
	titleDelimiter           = "-"
	titleRankStep            = "#"
	processedTitleBeginning  = "## "
	processedStepBeginning   = "- "
	stepIndicator            = "kdt.scenario.Step"
	stepDelimiter            = "`"
	stepPrefix               = "^"
	stepSuffix               = "$"
	methodPrefix             = ","
	methodSuffix             = ")"
	markdownCodeDelimiter    = "`"
	gherkinKeyword           = "<GK>"
	destinationFileBeginning = "# Syntax" + newLine + "Below you will find the step syntax next to the name of the method it utilizes. Here GK stands for [Gherkin](https://cucumber.io/docs/gherkin/reference/#keywords) Keyword and words in brackets ([]) are optional:" + newLine
)

var replacers = []struct {
	replacee string
	replacer string
}{
	{`(?:`, `[`},
	{` )?`, `] `},
	{`)?`, `]`},
	{`(\d+)`, `<digits>`},
	{`(\S+)`, `<non-whitespace-characters>`},
	{`([^"]*)`, `<any-characters-except-(")>`},
	{`\(`, `(`},
	{`\)`, `)`},
}

func main() {
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		log.Error(err)
	}
	defer sourceFile.Close()
	fileScanner := bufio.NewScanner(sourceFile)
	fileScanner.Split(bufio.ScanLines)
	rawSyntax := []string{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if isBeginAction(line) {
			log.Debugf("found begin action as '%s'", line)
			for fileScanner.Scan() {
				line := fileScanner.Text()
				if isEndAction(line) {
					log.Debugf("found end action as '%s'", line)
					break
				}
				rawSyntax = append(rawSyntax, line)
			}
			break
		}
	}
	log.Infof("found raw syntax to process as:")
	printStringSlice(rawSyntax)
	processedSyntax := processSyntax(rawSyntax)
	createSyntaxDocumentation(processedSyntax)
}

func createSyntaxDocumentation(processedSyntax []string) {
	if err := os.Remove(destinationFilePath); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(destinationFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing to '%s'", destinationFilePath)
	if _, err := f.WriteString(destinationFileBeginning); err != nil {
		log.Fatal(err)
	}
	for _, processedLine := range processedSyntax {
		if _, err := f.WriteString(processedLine); err != nil {
			log.Fatal(err)
		}
	}
	f.Close()
	printFile(destinationFilePath)
}

func processSyntax(rawSyntax []string) []string {
	processedSyntax := []string{}
	for _, rawLine := range rawSyntax {
		switch {
		case strings.Contains(rawLine, actionIndicator):
			title, titleRank := mustGetTitle(rawLine)
			titleBeginning := getTitleProcessedRank(titleRank)
			processedTitle := newLine + titleBeginning + title + newLine
			log.Debugf("processed '%s' as: '%s'", rawLine, processedTitle)
			processedSyntax = append(processedSyntax, processedTitle)
		case strings.Contains(rawLine, stepIndicator):
			processedStep := processedStepBeginning + processStep(rawLine) + newLine
			log.Debugf("processed '%s' as: '%s'", rawLine, processedStep)
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
	for _, r := range replacers {
		processedStep = strings.ReplaceAll(processedStep, r.replacee, r.replacer)
	}
	method := rawStepSplit[2]
	method = strings.TrimPrefix(method, methodPrefix)
	method = strings.TrimSuffix(method, methodSuffix)
	return markdownCodeDelimiter + gherkinKeyword + " " + processedStep + markdownCodeDelimiter + method
}

func mustGetTitle(line string) (string, int) {
	action, afterAction := getAction(line)
	actionSplit := strings.Split(action, titleDelimiter)
	if len(actionSplit) != 2 {
		log.Fatalf("expected '%s' to meet format '%s%s<digit>'", action, actionTitle, titleDelimiter)
	}
	action, titleRankString := actionSplit[0], actionSplit[1]
	if action != actionTitle {
		log.Fatalf("expected '%s' to contain '%s%s%s'", line, actionIndicator, actionDelimiter, actionTitle)
	}
	titleRank, err := strconv.Atoi(titleRankString)
	if err != nil {
		log.Fatalf("failed converting '%s' to integer: '%v'", titleRankString, err)
	}
	if afterAction == "" {
		log.Fatalf("expected '%s' to contain '%s%s%s%s<title>'", line, actionIndicator, actionDelimiter, actionTitle, actionDelimiter)
	}
	return afterAction, titleRank
}

func getTitleProcessedRank(rank int) string {
	if rank < 0 || rank > 9 {
		log.Fatalf("expected '%d' to be a digit between 1 and 9)", rank)
	}
	totalRankString := strings.Repeat(titleRankStep, rank)
	return totalRankString + processedTitleBeginning
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
		fmt.Println(s)
	}
}

func printFile(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed reading '%s': '%v'", path, err)
	}
	fmt.Println(string(file))
}
