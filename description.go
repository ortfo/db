package main

//TODO: handle files with *no language markers*!!!!!1
//TODO: deal with markdown extensions (see https://pkg.go.dev/github.com/gomarkdown/markdown/parser#Extensions):
// - french guillemets -> renderer:SmartypantsQuotesNBSP
// - open links in new tab -> renderer:HrefTargetBlank
// ...

//TODO: reorganize that file, it's a hot mess.

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	// "github.com/davecgh/go-spew/spew"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"

	// "github.com/gomarkdown/markdown/renderer"
	// "github.com/davecgh/go-spew/spew"
	"github.com/metal3d/go-slugify"
)

const (
	patternImageOrMediaOrLinkDeclaration string = `^([!>]?)\[([^"\]]+)(?: "([^"\]]+)")?\]\(([^\)]+)\)$`
	patternLanguageMarker                string = `^::\s+(.+)$`
	patternFootnoteDeclaration           string = `^\[\^([^\^]+)\]:\s+(.+)$`
	patternAbbreviationDefinition        string = `^\*\[([^\]]+)\]:\s+(.+)$`
	patternParagraphID                   string = `^\(([a-z-]+)\)$`
	patternTitle                         string = `^#\s+(.+)$`
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
	Name    string
	Content string
}

// Paragraph represents a paragraph declaration in a description.md file
type Paragraph struct {
	ID      string
	Content string
}

// Link represents an (isolated) link declaration in a description.md file
type Link struct {
	ID    string
	Name  string
	Title string
	URL   string
}

// Work represents a complete work, with analyzed mediae
type Work struct {
	Metadata   map[string]interface{}
	Title      map[string]string
	Paragraphs map[string][]Paragraph
	Media      map[string][]Media
	Links      map[string][]Link
	Footnotes  map[string][]Footnote
}

// ParsedDescription represents a work, but without analyzed media. All it contains is information from the description.md file
type ParsedDescription struct {
	Metadata               map[string]interface{}
	Title                  map[string]string
	Paragraphs             map[string][]Paragraph
	MediaEmbedDeclarations map[string][]MediaEmbedDeclaration
	Links                  map[string][]Link
	Footnotes              map[string][]Footnote
}

// Chunk binds some content to its chunk type
// Legal chunk types:
// - abbreviation
// - paragraphWithID
// - image
// - media
// - link
// - footnoteDeclaration
// - paragraph
// - title
type Chunk struct {
	Type    string
	Content string
}

// MediaEmbedDeclaration represents >[media](...) embeds.
// Only stores the info extracted from the syntax, no filesystem interactions.
type MediaEmbedDeclaration struct {
	Alt    string
	Title  string
	Source string
}

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

// ParseFootnote parses raw markdown into a footnote struct.
func ParseFootnote(markdownRaw string) Footnote {
	groups := RegexpGroups(patternFootnoteDeclaration, markdownRaw)
	return Footnote{Name: groups[1], Content: groups[2]}
}

// SplitOnLanguageMarkers returns two values:
// 1. the text before any language markers
// 2. a map with language codes as keys and the content as values
func SplitOnLanguageMarkers(markdownRaw string) (string, map[string]string) {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile(patternLanguageMarker)
	currentLanguage := ""
	before := ""
	markdownRawPerLanguage := map[string]string{}
	for _, line := range lines {
		if pattern.MatchString(line) {
			currentLanguage = pattern.FindStringSubmatch(line)[1]
			markdownRawPerLanguage[currentLanguage] = ""
		}
		if currentLanguage == "" {
			before += line + "\n"
		} else {
			markdownRawPerLanguage[currentLanguage] += line + "\n"
		}
	}
	return before, markdownRawPerLanguage
}

// ExtractTitle extracts the first <h1> from markdown
func ExtractTitle(line string) string {
	pattern := regexp.MustCompile(`^#\s+(.+)$`)
	if pattern.MatchString(line) {
		return pattern.FindStringSubmatch(line)[0]
	}
	return ""
}

// FindTitle searches through markdownRaw line-by-line until ExtractTitle finds title, or until it reaches the end.
func FindTitle(markdownRaw string) string {
	lines := strings.Split(markdownRaw, "\n")
	foundTitle := ""
	for _, line := range lines {
		foundTitle = line
	}
	return foundTitle
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
		ID:    slugify.Marshal(regexMatches[2]),
		Name:  regexMatches[2],
		Title: regexMatches[3],
		URL:   regexMatches[4],
	}
}

func extractMedia(regexMatches []string) MediaEmbedDeclaration {
	return MediaEmbedDeclaration{
		Alt:    regexMatches[2],
		Title:  regexMatches[3],
		Source: regexMatches[4],
	}
}

func extractAbbreviation(regexMatches []string) Abbreviation {
	return Abbreviation{
		Name:       regexMatches[1],
		Definition: regexMatches[2],
	}
}

// ParseParagraph takes a chunk of type "paragraph" or "paragraphWithID" and returns a parsed Paragraph with HTML content
func ParseParagraph(chunk Chunk) Paragraph {
	var paragraphID string = ""
	var paragraphContent string = chunk.Content
	if chunk.Type == "paragraphWithID" {
		paragraphID = RegexpGroups(patternParagraphID, strings.Split(chunk.Content, "\n")[0])[1]
		// Every line except the first (the paragraph id marker)
		paragraphContent = strings.Join(strings.Split(chunk.Content, "\n")[1:], "\n")
	}
	return Paragraph{
		Content: paragraphContent,
		ID:      paragraphID,
	}
}

// ParseLanguagedChunks takes in raw markdown without language markers (called on SplitOnLanguageMarker's output)
// and dispatches parsing to the appropriate functions, dependending on each chunk's type (a paragraph, an image, etc.)
func ParseLanguagedChunks(markdownRaw string) []Chunk {
	chunks := strings.Split(markdownRaw, "\n\n")
	typedChunks := make([]Chunk, 0)

	for _, chunk := range chunks {
		// Skip empty chunks
		chunk = strings.TrimSpace(chunk)
		if len(chunk) == 0 {
			continue
		} else if RegexpMatches(patternAbbreviationDefinition, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "abbreviation"})
		} else if RegexpMatches(patternParagraphID, strings.Split(chunk, "\n")[0]) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "paragraphWithID"})
		} else if RegexpMatches(patternImageOrMediaOrLinkDeclaration, chunk) {
			mediaOrImageOrLinkMarker := RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk)[1]
			if mediaOrImageOrLinkMarker == "" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "link"})
			} else if mediaOrImageOrLinkMarker == ">" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "media"})
			} else if mediaOrImageOrLinkMarker == "!" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "image"})
			}
		} else if RegexpMatches(patternFootnoteDeclaration, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "footnote"})
		} else if RegexpMatches(patternLanguageMarker, chunk) {
			continue
		} else if RegexpMatches(patternTitle, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "title"})
		} else {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "paragraph"})
		}
	}

	return typedChunks
}

// ParseMediaChunk takes a chunk of type "media" and returns a MediaEmbedDeclaration
func ParseMediaChunk(chunk Chunk) MediaEmbedDeclaration {
	return extractMedia(RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk.Content))
}

// ParseLinkChunk takes a chunk of type "link" and returns a Link
func ParseLinkChunk(chunk Chunk) Link {
	return extractLink(RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk.Content))
}

// ParseAbbreviationChunk takes a chunk of type "abbreviation" and returns an Abbreviation
func ParseAbbreviationChunk(chunk Chunk) Abbreviation {
	return extractAbbreviation(RegexpGroups(patternAbbreviationDefinition, chunk.Content))
}

// GetAllLanguages returns all language codes used in the document
func GetAllLanguages(markdownRaw string) []string {
	lines := strings.Split(markdownRaw, "\n")
	languages := make([]string, 0)
	for _, line := range lines {
		if RegexpMatches(patternLanguageMarker, line) {
			languages = append(languages, RegexpGroups(patternLanguageMarker, line)[1])
		}
	}
	return languages
}

// MarkdownToHTML converts markdown markdownRaw into an HTML string
func MarkdownToHTML(markdownRaw string) string {
	//TODO: handle markdown extensions (need to take in a "config Configuration" parameter)
	extensions := parser.CommonExtensions | parser.Footnotes | parser.AutoHeadingIDs
	return string(markdown.ToHTML([]byte(markdownRaw), parser.NewWithExtensions(extensions), nil))
}

// ProcessFootnoteReferences renders to HTML all footnote references in markdownRaw
func ProcessFootnoteReferences(markdownRaw string) string {
	patternFootnoteReference := regexp.MustCompile(`\[\^([^\]]+)\]`)
	processed := markdownRaw
	for _, referencePosition := range patternFootnoteReference.FindAllStringIndex(markdownRaw, -1) {
		reference := markdownRaw[referencePosition[0]:referencePosition[1]]
		footnoteName := patternFootnoteReference.FindStringSubmatch(reference)[1]
		//TODO: make the href (and <sup> class) customizable
		processed = strings.ReplaceAll(processed, reference, `<sup class="foonote-ref"><a href="#footnote:` + footnoteName + `">` + footnoteName + `</a></sup>`)
	}
	return processed
}

// ParseDescription parses the markdown string from a description.md file and returns a ParsedDescription
func ParseDescription(markdownRaw string) ParsedDescription {
	metadata, markdownRaw := ParseYAMLHeader(markdownRaw)
	// notLocalizedRaw: raw markdown before the first language marker
	notLocalizedRaw, localizedRawBlocks := SplitOnLanguageMarkers(markdownRaw)
	paragraphs := make(map[string][]Paragraph, 0)
	mediaEmbedDeclarations := make(map[string][]MediaEmbedDeclaration, 0)
	links := make(map[string][]Link, 0)
	title := make(map[string]string, 0)
	footnotes := make(map[string][]Footnote, 0)
	abbreviations := make(map[string][]Abbreviation, 0)
	// First pass to collect everything
	for _, language := range GetAllLanguages(markdownRaw) {
		// Unlocalized stuff appears the same in every language.
		chunks := ParseLanguagedChunks(notLocalizedRaw)
		chunks = append(chunks, ParseLanguagedChunks(localizedRawBlocks[language])...)
		currentLanguageParagraphs := make([]Paragraph, 0)
		currentLanguageMediaEmbedDeclarations := make([]MediaEmbedDeclaration, 0)
		currentLanguageLinks := make([]Link, 0)
		currentLanguageFootnotes := make([]Footnote, 0)
		currentLanguageAbbreviations := make([]Abbreviation, 0)
		var currentLanguageTitle string
		for _, chunk := range chunks {
			if chunk.Type == "title" {
				currentLanguageTitle = RegexpGroups(patternTitle, chunk.Content)[1]
			} else if chunk.Type == "footnote" {
				footnote := ParseFootnote(chunk.Content)
				currentLanguageFootnotes = append(currentLanguageFootnotes, footnote)
			} else if chunk.Type == "paragraph" || chunk.Type == "paragraphWithID" {
				currentLanguageParagraphs = append(currentLanguageParagraphs, ParseParagraph(chunk))
			} else if chunk.Type == "media" || chunk.Type == "image"{
				currentLanguageMediaEmbedDeclarations = append(currentLanguageMediaEmbedDeclarations, ParseMediaChunk(chunk))
			} else if chunk.Type == "links" {
				currentLanguageLinks = append(currentLanguageLinks, ParseLinkChunk(chunk))
			} else if chunk.Type == "abbreviation" {
				currentLanguageAbbreviations = append(currentLanguageAbbreviations, ParseAbbreviationChunk(chunk))
			}
		}
		// Second pass to replace abbreviations (if any), render footnote references (if any) and render to HTML
		for i, paragraph := range currentLanguageParagraphs {
			processed := paragraph.Content
			for _, abbreviation := range currentLanguageAbbreviations {
				var replacePattern = regexp.MustCompile(`\b` + abbreviation.Name + `\b`)
				processed = replacePattern.ReplaceAllString(paragraph.Content, "<abbr title=\""+abbreviation.Definition+"\">"+abbreviation.Name+"</abbr>")
			}
			processed = ProcessFootnoteReferences(processed)
			convertedToHTML := MarkdownToHTML(processed)
			// Fix footnote references' links having a number as text instead of the footnote's name: here, every paragraph is isolated, so the converter can't possibly asssign correct footnote reference numbers.
			// We use the footnote's name directly instead.
			// Remove outer paragraph tag & eventual whitespace
			patternOuterParagraph := `\s*<p>(.+)</p>\s*`
			convertedToHTML = RegexpGroups(patternOuterParagraph, convertedToHTML)[1]
			currentLanguageParagraphs[i] = Paragraph{
				ID:      paragraph.ID,
				Content: convertedToHTML,
			}
		}
		paragraphs[language] = currentLanguageParagraphs
		links[language] = currentLanguageLinks
		title[language] = currentLanguageTitle
		mediaEmbedDeclarations[language] = currentLanguageMediaEmbedDeclarations
		footnotes[language] = currentLanguageFootnotes
		abbreviations[language] = currentLanguageAbbreviations
	}
	return ParsedDescription{
		Metadata:               metadata,
		Paragraphs:             paragraphs,
		Links:                  links,
		Title:                  title,
		MediaEmbedDeclarations: mediaEmbedDeclarations,
		Footnotes:              footnotes,
	}
}
