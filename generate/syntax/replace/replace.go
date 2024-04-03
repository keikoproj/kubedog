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
		if c == ' ' {
			continue
		}
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
