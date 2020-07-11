package main

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	// "github.com/davecgh/go-spew/spew"
	// "github.com/gomarkdown/markdown"
	// "github.com/gomarkdown/markdown/parser"
	// "github.com/gomarkdown/markdown/renderer"
	"github.com/davecgh/go-spew/spew"
	"github.com/metal3d/go-slugify"
)

const (
	patternImageOrMediaOrLinkDeclaration string = `^([!>]?)\[([^"\]]+)(?: "([^"\]]+)")?\]\(([^\)]+)\)$`
	patternLanguageMarker                string = `^::\s+(.+)$`
	patternFootnoteDeclaration           string = `^\[(\d+)\]:\s+(.+)$`
	patternAbbreviationDefinition        string = `^\*\[([^\]]+)\]:\s+(.+)$`
	patternParagraphID                   string = `^\(([a-z-]+)\)$`
)

// ParseYAMLHeader parses the YAML header of a description markdown file and returns
// the rest of the content (all except the YAML header)
func ParseYAMLHeader(descriptionRaw string) (map[string]interface{}, string) {
	var inYAMLHeader bool
	var rawYAMLPart string
	var markdownPart string
	for _, line := range strings.Split(descriptionRaw, "\n") {
		// if strings.TrimSpace(line) == "" && !inYAMLHeader {
		// 	continue
		// }
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

// Abbreviation represents an abbreviation declaration in a description.md file
type Abbreviation struct {
	Name       string
	Definition string
}

// Footnote represents a footnote declaration in a description.md file
type Footnote struct {
	Number  uint16 // Oh no, what a bummer, you can't have more than 65 535 footnotes
	Content string
}

type Paragraph struct {
	ID      string
	Content string
}

type Link struct {
	ID    string
	Name  string
	Title string
	URL   string
}

type WorkObject struct {
	Name       string
	Paragraphs map[string][]Paragraph
	Media      map[string][]Media
	Links      map[string][]Link
	Colors     map[string]string
}

// MediaEmbedDeclaration represents >[media](...) embeds.
// Only stores the info extracted from the syntax, no filesystem interactions.
type MediaEmbedDeclaration struct {
	Alt    string
	Title  string
	Source string
}

// ImageEmbedDeclaration represents ![media](...) embeds.
// Only stores the info extracted from the syntax, no filesystem interactions.
type ImageEmbedDeclaration = MediaEmbedDeclaration

// CollectAbbreviation tries to match the given line and collect an abbreviation.
// Return values:
// 1. Abbreviation struct
// 2. Whether the line defines an abbreviation (bool)
func CollectAbbreviation(line string) (Abbreviation, bool) {
	pattern := regexp.MustCompile(patternAbbreviationDefinition)
	if pattern.MatchString(line) {
		matches := pattern.FindStringSubmatch(line)
		return Abbreviation{Name: matches[0], Definition: matches[1]}, true
	}
	return Abbreviation{}, false
}

// CollectFootnote tries to match the given line and collect a footnote declaration.
// Return values:
// 1. Footnote struct
// 2. Whether the line declares a footnote (bool)
func CollectFootnote(line string) (Footnote, bool) {
	pattern := regexp.MustCompile(patternFootnoteDeclaration)
	if pattern.MatchString(line) {
		matches := pattern.FindStringSubmatch(line)
		footnoteNumber, _ := strconv.ParseUint(matches[0], 10, 16)
		return Footnote{Number: uint16(footnoteNumber), Content: matches[1]}, true
	}
	return Footnote{}, false
}

// CollectAbbreviationsAndFootnotes iterates through the document's lines and
// extracts abbreviations and footnotes declarations from the file
// The first returned value is the markdown document with parsed declarations removed.
func CollectAbbreviationsAndFootnotes(markdownRaw string) (string, []Abbreviation, []Footnote) {
	lines := strings.Split(markdownRaw, "\n")
	markdownRet := ""
	abbreviations := make([]Abbreviation, 8^16)
	footnotes := make([]Footnote, 8^16)
	for _, line := range lines {
		abbreviation, definesAbbreviation := CollectAbbreviation(line)
		footnote, declaresFootnote := CollectFootnote(line)
		if definesAbbreviation {
			abbreviations = append(abbreviations, abbreviation)
		} else if declaresFootnote {
			footnotes = append(footnotes, footnote)
		} else {
			markdownRet += line + "\n"
			continue
		}

	}
	return markdownRet, abbreviations, footnotes
}

// SplitOnLanguageMarkers returns two values:
// 1. the text before any language markers
// 2. a map with language codes as keys and the content as values
func SplitOnLanguageMarkers(markdownRaw string) (string, map[string]string) {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile(patternLanguageMarker)
	currentLanguage := ""
	before := ""
	retMap := map[string]string{}
	for _, line := range lines {
		if pattern.MatchString(line) {
			currentLanguage := pattern.FindStringSubmatch(line)[0]
			retMap[currentLanguage] = ""
		}
		if currentLanguage == "" {
			before += line + "\n"
		} else {
			retMap[currentLanguage] += line + "\n"
		}
	}
	return before, retMap
}

// ExtractName extracts the first <h1> from markdown
func ExtractName(line string) string {
	pattern := regexp.MustCompile(`^#\s+(.+)$`)
	if pattern.MatchString(line) {
		return pattern.FindStringSubmatch(line)[0]
	}
	return ""
}

// ExtractMedia extracts media declarations (>[alt "title"](source)), images (![alt "title"](source)) or links ([alt "title"](source))
// Return value is a regex match string array: first character (empty for links), alt, title, source.
func extractMediaOrImageOrLink(line string) []string {
	pattern := regexp.MustCompile(patternImageOrMediaOrLinkDeclaration)
	if pattern.MatchString(line) {
		matches := pattern.FindStringSubmatch(line)
		return matches
	}
	return make([]string, 0)
}

func extractLink(regexMatches []string) Link {
	return Link{
		ID:    slugify.Marshal(regexMatches[1]),
		Name:  regexMatches[1],
		Title: regexMatches[2],
		URL:   regexMatches[3],
	}
}

func extractImage(regexMatches []string) ImageEmbedDeclaration {
	return ImageEmbedDeclaration{
		Alt:    regexMatches[1],
		Title:  regexMatches[2],
		Source: regexMatches[3],
	}
}

func extractMedia(regexMatches []string) MediaEmbedDeclaration {
	return MediaEmbedDeclaration{
		Alt:    regexMatches[1],
		Title:  regexMatches[2],
		Source: regexMatches[3],
	}
}

// ExtractParagraphs extracts the paragraphs and their IDs (if present)
// and returns an array of paragraphs
func ExtractParagraphs(markdownRaw string) []Paragraph {
	chunks := strings.Split(markdownRaw, "\n\n")
	spew.Dump(chunks)
	paragraphs := make([]Paragraph, 0)
	firstParagraphIDPattern := regexp.MustCompile(patternParagraphID)
	currentParagraphID := ""

	for _, chunk := range chunks {
		chunkLines := strings.Split(chunk, "\n")
		isAParagraphChunk := !RegexpMatches(patternAbbreviationDefinition, chunk) && !RegexpMatches(patternFootnoteDeclaration, chunk) && !RegexpMatches(patternImageOrMediaOrLinkDeclaration, chunk) && !RegexpMatches(patternLanguageMarker, chunk)

		if firstParagraphIDPattern.MatchString(chunkLines[0]) {
			currentParagraphID = firstParagraphIDPattern.FindStringSubmatch(chunkLines[0])[0]
			chunkLines = chunkLines[1:]
		}
		if isAParagraphChunk {
			paragraphs = append(paragraphs, Paragraph{
				ID:      currentParagraphID,
				Content: strings.Join(chunkLines, "\n"),
			})
		}
	}
	return paragraphs
}
