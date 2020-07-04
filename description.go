package main

import (
	// "fmt"
	"regexp"
	"strings"

	// "github.com/davecgh/go-spew/spew"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v2"
)

// UnknownYAMLObject represents an object with unknown structure
// and is used to store the YAML header of description.md files
type UnknownYAMLObject interface{}

// DescriptionParseResult represents a parsed description.md file.
// It does *not* represent a complete project object, though, as
// it is not aware of any media files in the project's directory,
// for example.
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
// and returns them as a map, with keys being the abbreivations and values their respective definitions as the first return value
// and the raw markdown string with abbreviation declarations stripped.
func CollectAbbreviationDeclarations(markdownRaw string) (map[string]string, string) {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile("\\s*\\*\\[([^\\]]+)\\]: (.+)")
	abbreviations := make(map[string]string)
	var markdownStripped string
	for _, line := range lines {
		isAnAbbreviationDefinition := pattern.MatchString(line)
		if isAnAbbreviationDefinition {
			groups := pattern.FindStringSubmatch(line)
			abbreviations[groups[1]] = groups[2]
		} else {
			markdownStripped += line + "\n"
		}
	}
	return abbreviations, markdownStripped
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
