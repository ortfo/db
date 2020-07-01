package main

import (
	"strings"
	"regexp"
	"fmt"
	
	"gopkg.in/yaml.v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

type UnknownYAMLObject interface{}

type DescriptionParseResult struct {
	HTML string
	YAMLHeader map[string]interface{}
	MadeWith []string
	Links map[string]string
	Tags []string
}

// ParseYAMLHeader parses the YAML header of a description markdown file and returns
// the rest of the content (all except the YAML header)
func ParseYAMLHeader(descriptionRaw string) (map[string]interface{}, string) {
	var inYAMLHeader bool
	var rawYAMLPart string
	var markdownPart string
	for _, line := range strings.Split(descriptionRaw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.TrimSpace(line) == "---" {
			inYAMLHeader = !inYAMLHeader
			continue
		}
		if inYAMLHeader {
			rawYAMLPart += line + "\n"
		} else {
			markdownPart += line + "\n"
		}
	}
	var parsedYAMLPart map[string]interface{}
	yaml.Unmarshal([]byte(rawYAMLPart), &parsedYAMLPart)
	return parsedYAMLPart, markdownPart
}

func ConvertMarkdownToHTML(markdownRaw string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)
	markdownBytes := []byte(markdownRaw)
	return string(markdown.ToHTML(markdownBytes, parser, nil))
}

func ParseCompactDescriptionList(markdownRaw string) string {
	lines := strings.Split(markdownRaw, "\n")
	// var parsedLines []string
	pattern, _ := regexp.Compile("(\\s*)- +([^:]+): (.+)")
	for _, line := range lines {
		groups := pattern.FindString(line)
		fmt.Println(groups)
	}
	return "heh"
}

func CollectAbbreviationDeclarations(markdownRaw string) map[string]string {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile("\\s*\\[\\*(?P<Abbreviation>[^\\]]+)\\]: (?P<Definition>.+)")
	var abbreviations map[string]string
	for _, line := range lines {
		fmt.Printf("%#v\n", pattern.FindStringSubmatch(line))
		groups := pattern.SubexpNames()
		fmt.Printf("%#v\n", groups)
		abbreviations[groups[0]] = groups[1]
	}
	return abbreviations
}

func ReplaceAbbreviations(markdownRaw string, abbreviations map[string]string) string {
	for abbr, def := range abbreviations {
		markup := "<abbr title=\"" + def + "\">" + abbr + "</abbr>"
		markdownRaw = strings.ReplaceAll(markdownRaw, abbr, markup)
	}
	return markdownRaw
}
