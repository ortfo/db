package main

import (
	// "fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v2"
)

type UnknownYAMLObject interface{}

type DescriptionParseResult struct {
	HTML       string
	YAMLHeader map[string]interface{}
	MadeWith   []string
	Links      map[string]string
	Tags       []string
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

// ConvertMarkdownToHTML converts a markdown string to an HTML string,
// using CommonExtensions and AutoHeadingIDs extensions github.com/gomarkdown/markdown
func ConvertMarkdownToHTML(markdownRaw string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)
	markdownBytes := []byte(markdownRaw)
	return string(markdown.ToHTML(markdownBytes, parser, nil))
}

// CollectAbbreviationDeclarations looks for Abbreviations & acronyms definitions in a markdown string
// and returns them as a map, with keys being the abbreivations and values their respective definitions
func CollectAbbreviationDeclarations(markdownRaw string) map[string]string {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile("\\s*\\*\\[([^\\]]+)\\]: (.+)")
	abbreviations := make(map[string]string)
	for _, line := range lines {
		isAnAbbreviationDefinition := pattern.MatchString(line)
		if isAnAbbreviationDefinition {
			groups := pattern.FindStringSubmatch(line)
			abbreviations[groups[1]] = groups[2]
		}
	}
	return abbreviations
}

// ReplaceAbbreviations takes in a markdown string and a map of abbreviation: definition and replaces
// occurences of ``abbreviation`` with the appropriate HTML markup (<abbr> tag)
func ReplaceAbbreviations(markdownRaw string, abbreviations map[string]string) string {
	for abbr, def := range abbreviations {
		//TODO: Replace on word boundaries
		escapedDef := strings.ReplaceAll(def, "\"", "\\\"")
		markup := "<abbr title=\"" + escapedDef + "\">" + abbr + "</abbr>"
		markdownRaw = strings.ReplaceAll(markdownRaw, abbr, markup)
	}
	return markdownRaw
}
