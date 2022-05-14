package ortfodb

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v2"

	"github.com/anaskhan96/soup"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"

	"github.com/metal3d/go-slugify"
)

const (
	PatternLanguageMarker         string = `^::\s+(.+)$`
	PatternAbbreviationDefinition string = `^\s*\*\[([^\]]+)\]:\s+(.+)$`
	RuneLooped                    rune   = '~'
	RuneAutoplay                  rune   = '>'
	RuneHideControls              rune   = '='
)

// ParseYAMLHeader parses the YAML header of a description markdown file and returns the rest of the content (all except the YAML header).
func ParseYAMLHeader(descriptionRaw string) (map[string]interface{}, string) {
	var inYAMLHeader bool
	var rawYAMLPart string
	var markdownPart string
	for _, line := range strings.Split(descriptionRaw, "\n") {
		// Replace tabs with four spaces
		for strings.HasPrefix(line, "\t") {
			line = strings.Repeat(" ", 4) + strings.TrimPrefix(line, "\t")
		}
		// A YAML header separator is 3 or more dashes on a line (without anything else on the same line)
		if strings.Trim(line, "-") == "" && strings.Count(line, "-") >= 3 {
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
	if parsedYAMLPart == nil {
		parsedYAMLPart = make(map[string]interface{})
	}
	return parsedYAMLPart, markdownPart
}

// ParseDescription parses the markdown string from a description.md file and returns a ParsedDescription.
func (ctx *RunContext) ParseDescription(markdownRaw string) ParsedDescription {
	metadata, markdownRaw := ParseYAMLHeader(markdownRaw)
	// notLocalizedRaw: raw markdown before the first language marker
	notLocalizedRaw, localizedRawBlocks := SplitOnLanguageMarkers(markdownRaw)
	localized := len(localizedRawBlocks) > 0
	var allLanguages []string
	if localized {
		allLanguages = mapKeys(localizedRawBlocks)
	} else {
		allLanguages = make([]string, 1)
		allLanguages[0] = "default" // TODO: make this configurable
	}
	paragraphs := make(map[string][]Paragraph)
	mediaEmbedDeclarations := make(map[string][]MediaEmbedDeclaration)
	links := make(map[string][]Link)
	title := make(map[string]string)
	footnotes := make(map[string]Footnotes)
	abbreviations := make(map[string]Abbreviations)
	for _, language := range allLanguages {
		// Unlocalized stuff appears the same in every language.
		raw := notLocalizedRaw
		if localized {
			raw += localizedRawBlocks[language]
		}
		title[language], paragraphs[language], mediaEmbedDeclarations[language], links[language], footnotes[language], abbreviations[language] = ParseSingleLanguageDescription(raw)
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

// Abbreviations represents the abbreviations declared in a description.md file.
type Abbreviations map[string]string

// Footnotes represents the footnote declarations in a description.md file.
type Footnotes map[string]string

// Paragraph represents a paragraph declaration in a description.md file.
type Paragraph struct {
	ID      string
	Content string
}

// Link represents an (isolated) link declaration in a description.md file.
type Link struct {
	ID    string
	Name  string
	Title string
	URL   string
}

// Work represents a complete work, with analyzed mediae.
type Work struct {
	ID         string
	Metadata   map[string]interface{}
	Title      map[string]string
	Paragraphs map[string][]Paragraph
	Media      map[string][]Media
	Links      map[string][]Link
	Footnotes  map[string]Footnotes
}

// MediaEmbedDeclaration represents media embeds. (abusing the ![]() syntax to extend it to any file).
// Only stores the info extracted from the syntax, no filesystem interactions.
type MediaEmbedDeclaration struct {
	Alt        string
	Title      string
	Source     string
	Attributes MediaAttributes
}

// MediaAttributes stores which HTML attributes should be added to the media.
type MediaAttributes struct {
	Looped      bool // Controlled with attribute character ~ (adds)
	Autoplay    bool // Controlled with attribute character > (adds)
	Muted       bool // Controlled with attribute character > (adds)
	Playsinline bool // Controlled with attribute character = (adds)
	Controls    bool // Controlled with attribute character = (removes)
}

// ParsedDescription represents a work, but without analyzed media. All it contains is information from the description.md file.
type ParsedDescription struct {
	Metadata               map[string]interface{}
	Title                  map[string]string
	Paragraphs             map[string][]Paragraph
	MediaEmbedDeclarations map[string][]MediaEmbedDeclaration
	Links                  map[string][]Link
	Footnotes              map[string]Footnotes
}

// SplitOnLanguageMarkers returns two values:
// 		1. the text before any language markers
// 		2. a map with language codes as keys and the content as values.
func SplitOnLanguageMarkers(markdownRaw string) (string, map[string]string) {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile(PatternLanguageMarker)
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

// ParseSingleLanguageDescription takes in raw markdown without language markers (called on splitOnLanguageMarker's output).
// and returns parsed arrays of structs that make up each language's part in ParsedDescription's maps.
func ParseSingleLanguageDescription(markdownRaw string) (title string, paragraphs []Paragraph, mediae []MediaEmbedDeclaration, links []Link, footnotes Footnotes, abbreviations Abbreviations) {
	markdownRaw = HandleAltMediaEmbedSyntax(markdownRaw)
	htmlRaw := MarkdownToHTML(markdownRaw)
	htmlTree := soup.HTMLParse(htmlRaw)
	paragraphs = make([]Paragraph, 0)
	mediae = make([]MediaEmbedDeclaration, 0)
	links = make([]Link, 0)
	footnotes = make(Footnotes)
	abbreviations = make(Abbreviations)
	paragraphLike := make([]soup.Root, 0)
	paragraphLikeTagNames := "p ol ul h2 h3 h4 h5 h6 dl blockquote hr pre"
	for _, element := range htmlTree.Find("body").Children() {
		// Check if it's a paragraph-like tag
		if strings.Contains(paragraphLikeTagNames, element.NodeValue) {
			paragraphLike = append(paragraphLike, element)
		}
	}
	for _, paragraph := range paragraphLike {
		childrenCount := len(paragraph.Children())
		firstChild := soup.Root{}
		if childrenCount >= 1 {
			firstChild = paragraph.Children()[0]
		}
		if childrenCount == 1 && firstChild.NodeValue == "img" {
			// A media embed
			alt, attributes := ExtractAttributesFromAlt(firstChild.Attrs()["alt"])
			mediae = append(mediae, MediaEmbedDeclaration{
				Alt:        alt,
				Title:      firstChild.Attrs()["title"],
				Source:     firstChild.Attrs()["src"],
				Attributes: attributes,
			})
		} else if childrenCount == 1 && firstChild.NodeValue == "a" {
			// An isolated link
			links = append(links, Link{
				ID:    slugify.Marshal(firstChild.FullText(), true),
				Name:  innerHTML(firstChild),
				Title: firstChild.Attrs()["title"],
				URL:   firstChild.Attrs()["href"],
			})
		} else if regexpMatches(PatternAbbreviationDefinition, innerHTML(paragraph)) {
			// An abbreviation definition
			groups := regexpGroups(PatternAbbreviationDefinition, innerHTML(paragraph))
			abbreviations[groups[1]] = groups[2]
		} else if regexpMatches(PatternLanguageMarker, innerHTML(paragraph)) {
			// A language marker (ignored)
			continue
		} else {
			// A paragraph (anything else)
			paragraphs = append(paragraphs, Paragraph{
				ID:      paragraph.Attrs()["id"],
				Content: paragraph.HTML(),
			})
		}
	}
	if h1 := htmlTree.Find("h1"); h1.Error == nil {
		title = innerHTML(h1)
		for _, div := range htmlTree.FindAll("div") {
			if div.Attrs()["class"] == "footnotes" {
				for _, li := range div.FindAll("li") {
					footnotes[strings.TrimPrefix(li.Attrs()["id"], "fn:")] = trimHTMLWhitespace(innerHTML(li))
				}
			}
		}
	}
	processedParagraphs := make([]Paragraph, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if strings.HasPrefix(paragraph.Content, "<pre>") && strings.HasSuffix(paragraph.Content, "</pre>") {
			// Dont insert <abbr>s while in <pre> text
			continue
		}
		processedParagraphs = append(processedParagraphs, ReplaceAbbreviations(paragraph, abbreviations))
	}
	return title, processedParagraphs, mediae, links, footnotes, abbreviations
}

// trimHTMLWhitespace removes whitespace from the beginning and end of an HTML string, also removing leading & trailing <br> tags.
func trimHTMLWhitespace(rawHTML string) string {
	rawHTML = strings.TrimSpace(rawHTML)
	for _, toRemove := range []string{"<br>", "<br />", "<br/>"} {
		for strings.HasPrefix(rawHTML, toRemove) {
			rawHTML = strings.TrimPrefix(rawHTML, toRemove)
		}
		for strings.HasSuffix(rawHTML, toRemove) {
			rawHTML = strings.TrimSuffix(rawHTML, toRemove)
		}
	}
	return rawHTML
}

// HandleAltMediaEmbedSyntax handles the >[...](...) syntax by replacing it in htmlRaw with ![...](...).
func HandleAltMediaEmbedSyntax(markdownRaw string) string {
	pattern := regexp.MustCompile(`(?m)^>(\[[^\]]+\]\([^)]+\)\s*)$`)
	return pattern.ReplaceAllString(markdownRaw, "!$1")
}
// ExtractAttributesFromAlt extracts sigils from the end of the alt attribute, returns the alt without them as well as the parse result.
func ExtractAttributesFromAlt(alt string) (string, MediaAttributes) {
	attrs := MediaAttributes{
		Controls: true, // Controls is added by default, others aren't
	}
	lastRune, _ := utf8.DecodeLastRuneInString(alt)
	// If there are no attributes in the alt string, the first (last in the alt string) will not be an attribute character.
	if !isMediaEmbedAttribute(lastRune) {
		return alt, attrs
	}
	returnedAlt := ""
	// We iterate backwards:
	// if there are attributes, they'll be at the end of the alt text separated by a space
	inAttributesZone := true
	for i := len([]rune(alt)) - 1; i >= 0; i-- {
		revChar := []rune(alt)[i]
		if revChar == ' ' && inAttributesZone {
			inAttributesZone = false
			continue
		}
		if inAttributesZone {
			if revChar == RuneAutoplay {
				attrs.Autoplay = true
				attrs.Muted = true
			} else if revChar == RuneLooped {
				attrs.Looped = true
			} else if revChar == RuneHideControls {
				attrs.Controls = false
				attrs.Playsinline = true
			}
		} else {
			// TODO better variable name
			returnedAlt = string(revChar) + returnedAlt
		}
	}
	return returnedAlt, attrs
}

func isMediaEmbedAttribute(char rune) bool {
	return char == RuneAutoplay || char == RuneLooped || char == RuneHideControls
}

// innerHTML returns the HTML string of what's _inside_ the given element, just like JS' `element.innerHTML`.
func innerHTML(element soup.Root) string {
	var innerHTML string
	for _, child := range element.Children() {
		innerHTML += child.HTML()
	}
	if innerHTML == "" {
		innerHTML = element.HTML()
	}
	return innerHTML
}

// MarkdownToHTML converts markdown markdownRaw into an HTML string.
func MarkdownToHTML(markdownRaw string) string {
	// TODO: add (ctx *RunContext) receiver, take markdown configuration into account when activating extensions
	extensions := parser.CommonExtensions | // Common stuff
		parser.Footnotes | // [^1]: footnotes
		parser.AutoHeadingIDs | // Auto-add [id] to headings
		parser.Attributes | // Specify attributes manually with {} above block
		parser.HardLineBreak | // \n becomes <br>
		parser.OrderedListStart | // Starting an <ol> with 5. will make them start at 5 in the output HTML
		parser.EmptyLinesBreakList // 2 empty lines break out of list
		// TODO: smart fractions, LaTeX-style dash parsing, smart quotes (see https://pkg.go.dev/github.com/gomarkdown/markdown@v0.0.0-20210514010506-3b9f47219fe7#readme-extensions)

	return string(markdown.ToHTML([]byte(markdownRaw), parser.NewWithExtensions(extensions), nil))
}

// ReplaceAbbreviations processes the given Paragraph to replace abbreviations.
func ReplaceAbbreviations(paragraph Paragraph, currentLanguageAbbreviations Abbreviations) Paragraph {
	processed := paragraph.Content
	for name, definition := range currentLanguageAbbreviations {
		var replacePattern = regexp.MustCompile(`\b` + name + `\b`)
		processed = replacePattern.ReplaceAllString(paragraph.Content, "<abbr title=\""+definition+"\">"+name+"</abbr>")
	}

	return Paragraph{Content: processed}
}
