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

package replace

import (
	"bytes"
	"regexp"
	"strings"
)

const regExp_CharsWithinBrackets = "([^(]*)"

type Replacement struct {
	Replacee string
	Replacer string
}

func (r Replacement) Replace(src string) string {
	return strings.ReplaceAll(src, r.Replacee, r.Replacer)
}

type Replacements []Replacement

func (rs Replacements) Replace(src string) string {
	new := src
	for _, r := range rs {
		new = r.Replace(new)
	}
	return new
}

type BracketsReplacement struct {
	Opening Replacement
	Closing Replacement
}

func (br BracketsReplacement) Replace(src string) string {
	re, _ := regexp.Compile(br.getRegExp())
	// TODO: check error
	new := re.ReplaceAllFunc([]byte(src), br.replaceSingle)
	return string(new)
}

func (br BracketsReplacement) replaceSingle(src []byte) []byte {
	s := string(src)
	s = br.Opening.Replace(s)
	s = br.Closing.Replace(s)
	return []byte(s)
}

func (br BracketsReplacement) getRegExp() string {
	return escapeEveryCharacter(br.Opening.Replacee) +
		regExp_CharsWithinBrackets +
		escapeEveryCharacter(br.Closing.Replacee)
}

func escapeEveryCharacter(s string) string {
	var buffer bytes.Buffer
	for _, c := range s {
		buffer.WriteString(`\`)
		buffer.WriteRune(c)
	}
	return buffer.String()
}

type BracketsReplacements []BracketsReplacement

func (brs BracketsReplacements) Replace(src string) string {
	new := src
	for _, br := range brs {
		new = br.Replace(new)
	}
	return new
}
