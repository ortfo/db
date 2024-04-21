package ortfodb

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v2"

	"github.com/anaskhan96/soup"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/k3a/html2text"
	"github.com/metal3d/go-slugify"
	"github.com/mitchellh/mapstructure"
	"github.com/relvacode/iso8601"
)

const (
	PatternLanguageMarker         string = `^::\s+(.+)$`
	PatternAbbreviationDefinition string = `^\s*\*\[([^\]]+)\]:\s+(.+)$`
	RuneLoop                      rune   = '~'
	RuneAutoplay                  rune   = '>'
	RuneHideControls              rune   = '='
)

// ParseYAMLHeader parses the YAML header of a description markdown file and returns the rest of the content (all except the YAML header).
func ParseYAMLHeader[Metadata interface{}](descriptionRaw string) (Metadata, string) {
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

	var metadata Metadata
	for key, value := range parsedYAMLPart {
		if strings.Contains(key, " ") {
			parsedYAMLPart[strings.ReplaceAll(key, " ", "_")] = value
			delete(parsedYAMLPart, key)
		}
	}
	mapstructure.Decode(parsedYAMLPart, &metadata)
	return metadata, markdownPart
}

// ParseDescription parses the markdown string from a description.md file.
// Media content blocks are left unanalyzed.
// BuiltAt and DescriptionHash are also not set.
func ParseDescription(ctx *RunContext, markdownRaw string, workID string) (Work, error) {
	defer TimeTrack(time.Now(), "ParseDescription", workID)
	metadata, markdownRaw := ParseYAMLHeader[WorkMetadata](markdownRaw)
	// notLocalizedRaw: raw markdown before the first language marker
	notLocalizedRaw, localizedRawBlocks := SplitOnLanguageMarkers(markdownRaw)
	LogDebug("split description into notLocalizedRaw: %#v and localizedRawBlocks: %#v", notLocalizedRaw, localizedRawBlocks)
	localized := len(localizedRawBlocks) > 0
	var allLanguages []string
	if localized {
		allLanguages = mapKeys(localizedRawBlocks)
	} else {
		// TODO: make this configurable
		allLanguages = []string{"default"}
	}
	contentsPerLanguage := LocalizableContent{}
	for _, language := range allLanguages {
		// Unlocalized stuff appears the same in every language.
		raw := notLocalizedRaw
		if localized {
			raw += localizedRawBlocks[language]
		}

		content := LocalizedContent{}

		content.Title, content.Blocks, content.Footnotes, content.Abbreviations = ctx.ParseSingleLanguageDescription(raw)
		var err error
		content.Layout, err = ResolveLayout(metadata, language, content.Blocks)
		if err != nil {
			return Work{}, fmt.Errorf("while resolving %s layout: %w", language, err)
		}

		contentsPerLanguage[language] = content
	}

	return Work{
		ID:       workID,
		Content:  contentsPerLanguage,
		Metadata: metadata,
	}, nil
}

// Abbreviations represents the abbreviations declared in a description.md file.
type Abbreviations map[string]string

// Footnotes represents the footnote declarations in a description.md file.
type Footnotes map[string]HTMLString

// Paragraph represents a paragraph declaration in a description.md file.
type Paragraph struct {
	Content HTMLString `json:"content"` // html
}

// Link represents an (isolated) link declaration in a description.md file.
type Link struct {
	Text  HTMLString `json:"text"`
	Title string     `json:"title"`
	URL   string     `json:"url"`
}

// Work represents a given work in the database. It may or not have analyzed media.
type Work struct {
	ID              string             `json:"id"`
	BuiltAt         time.Time          `json:"builtAt"`
	DescriptionHash string             `json:"descriptionHash"`
	Metadata        WorkMetadata       `json:"metadata"`
	Content         LocalizableContent `json:"content"`
	Partial         bool               `json:"Partial"`
}

func (w Work) ThumbnailBlock(language string) Media {
	firstMatch := Media{}
	for _, block := range w.Content.Localize(language).Blocks {
		if !block.Type.IsMedia() {
			continue
		}

		if firstMatch.DistSource == "" {
			firstMatch = block.Media
		}

		if block.Media.RelativeSource == w.Metadata.Thumbnail {
			return block.Media
		}
	}
	return firstMatch
}

func (w Work) ThumbnailPath(language string, size int) FilePathInsideMediaRoot {
	return w.ThumbnailBlock(language).Thumbnails.Closest(size)
}

func (w Work) Colors(language string) ColorPalette {
	if !w.Metadata.Colors.Empty() {
		return w.Metadata.Colors
	}

	thumb := w.ThumbnailBlock(language)

	if !thumb.Colors.Empty() {
		return thumb.Colors
	}

	for _, block := range w.Content[language].Blocks {
		if !block.Type.IsMedia() {
			continue
		}
		if block.AsMedia().Colors.Empty() {
			continue
		}
		return block.AsMedia().Colors
	}

	return ColorPalette{}
}

func (thumbnails ThumbnailsMap) Closest(size int) FilePathInsideMediaRoot {
	if len(thumbnails) == 0 {
		return ""
	}
	var closest int
	for thumbnailSize := range thumbnails {
		if thumbnailSize > closest && thumbnailSize <= size {
			closest = thumbnailSize
		}
	}
	return thumbnails[closest]
}

type WorkMetadata struct {
	Aliases            []string                      `json:"aliases" yaml:",omitempty"`
	Finished           string                        `json:"finished" yaml:",omitempty"`
	Started            string                        `json:"started"`
	MadeWith           []string                      `json:"madeWith" yaml:"made with"`
	Tags               []string                      `json:"tags"`
	Thumbnail          FilePathInsidePortfolioFolder `json:"thumbnail" yaml:",omitempty"`
	TitleStyle         TitleStyle                    `json:"titleStyle" yaml:"title style,omitempty"`
	Colors             ColorPalette                  `json:"colors" yaml:",omitempty"`
	PageBackground     string                        `json:"pageBackground" yaml:"page background,omitempty"`
	WIP                bool                          `json:"wip" yaml:",omitempty"`
	Private            bool                          `json:"private" yaml:",omitempty"`
	AdditionalMetadata map[string]interface{}        `mapstructure:",remain" json:"additionalMetadata" yaml:",omitempty"`
	DatabaseMetadata   DatabaseMeta                  `json:"databaseMetadata" yaml:"-" `
}

func (m WorkMetadata) CreatedAt() time.Time {
	var creationDate string
	if m.AdditionalMetadata["created"] != nil {
		creationDate = m.AdditionalMetadata["created"].(string)
	} else if m.Finished != "" {
		creationDate = m.Finished
	} else {
		creationDate = m.Started
	}
	if creationDate == "" {
		return time.Date(9999, time.January, 1, 0, 0, 0, 0, time.Local)
	}
	parsedDate, err := parsePossiblyInterderminateDate(creationDate)
	if err != nil {
		panic(err)
	}
	return parsedDate
}

func parsePossiblyInterderminateDate(datestring string) (time.Time, error) {
	return iso8601.ParseString(
		strings.ReplaceAll(
			strings.Replace(datestring, "????", "9999", 1), "?", "1",
		),
	)
}

type TitleStyle string

type LocalizableContent map[string]LocalizedContent

func (c LocalizableContent) Localize(lang string) LocalizedContent {
	if len(c) == 0 {
		return LocalizedContent{}
	}

	if _, ok := c[lang]; ok {
		return c[lang]
	}

	return c["default"]
}

type LocalizedContent struct {
	Layout        Layout         `json:"layout"`
	Blocks        []ContentBlock `json:"blocks"`
	Title         HTMLString     `json:"title"`
	Footnotes     Footnotes      `json:"footnotes"`
	Abbreviations Abbreviations  `json:"abbreviations"`
}

type ContentBlock struct {
	ID     string           `json:"id"`
	Type   ContentBlockType `json:"type"`
	Anchor string           `json:"anchor"`
	Index  int              `json:"index"`
	Media
	Paragraph
	Link
}

func (b ContentBlock) AsMedia() Media {
	if b.Type != "media" {
		panic("ContentBlock is not a media")
	}

	return Media{
		Alt:            b.Alt,
		Caption:        b.Caption,
		DistSource:     b.DistSource,
		RelativeSource: b.RelativeSource,
		ContentType:    b.ContentType,
		Size:           b.Size,
		Dimensions:     b.Dimensions,
		Online:         b.Online,
		Duration:       b.Duration,
		Colors:         b.Colors,
		Thumbnails:     b.Thumbnails,
		Attributes:     b.Attributes,
	}
}

func (b ContentBlock) AsLink() Link {
	if b.Type != "link" {
		panic("ContentBlock is not a link")
	}

	return Link{
		Text:  b.Text,
		Title: b.Link.Title,
		URL:   b.URL,
	}
}

func (b ContentBlock) AsParagraph() Paragraph {
	if b.Type != "paragraph" {
		panic("ContentBlock is not a paragraph")
	}

	return Paragraph{
		Content: b.Content,
	}
}

type ThumbnailsMap map[int]FilePathInsideMediaRoot

// FilePathInsidePortfolioFolder is a path relative to the scattered mode folder inside of a work directory. (example ../image.jpeg for an image in the work's directory, just outside of the portfolio-specific folder)
type FilePathInsidePortfolioFolder string

// FilePathInsideMediaRoot is a path relative to the media root directory.
type FilePathInsideMediaRoot string

func (f FilePathInsidePortfolioFolder) Absolute(ctx *RunContext, workID string) string {
	result, _ := filepath.Abs(filepath.Join(ctx.DatabaseDirectory, workID, ctx.Config.ScatteredModeFolder, string(f)))
	return result
}

func (f FilePathInsideMediaRoot) URL(origin string) string {
	return origin + "/" + string(f)
}

type HTMLString string

func (s HTMLString) String() string {
	return html2text.HTML2Text(string(s))
}

// ContentBlockType is one of "paragraph", "media" or "link"
type ContentBlockType string

func (t ContentBlockType) String() string {
	return string(t)
}

func (t ContentBlockType) IsParagraph() bool {
	return string(t) == "paragraph"
}

func (t ContentBlockType) IsMedia() bool {
	return string(t) == "media"
}

func (t ContentBlockType) IsLink() bool {
	return string(t) == "link"
}

// Layout is a 2D array of content block IDs
type Layout [][]LayoutCell

// LayoutCell is a single cell in the layout. It corresponds to the content block's ID.
type LayoutCell string

// MediaAttributes stores which HTML attributes should be added to the media.
type MediaAttributes struct {
	Loop        bool `json:"loop"`        // Controlled with attribute character ~ (adds)
	Autoplay    bool `json:"autoplay"`    // Controlled with attribute character > (adds)
	Muted       bool `json:"muted"`       // Controlled with attribute character > (adds)
	Playsinline bool `json:"playsinline"` // Controlled with attribute character = (adds)
	Controls    bool `json:"controls"`    // Controlled with attribute character = (removes)
}

// ParsedWork represents a work, but without analyzed media. All it contains is information from the description.md file.
type ParsedWork Work

// SplitOnLanguageMarkers returns two values:
//  1. the text before any language markers
//  2. a map with language codes as keys and the content as values.
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

// generatedID returns the ID of the content block. It is only as unique as the data it is based on.
func (b ContentBlock) generateID() string {
	var dataToUse string
	switch b.Type {
	case "media":
		dataToUse = string(b.Media.RelativeSource)
	case "paragraph":
		dataToUse = string(b.AsParagraph().Content)
	case "link":
		dataToUse = b.Link.URL
	}
	hash := md5.Sum([]byte(string(b.Type) + dataToUse))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])[:10]
}

// ParseSingleLanguageDescription takes in raw markdown without language markers (called on splitOnLanguageMarker's output).
// and returns parsed arrays of structs that make up each language's part in ParsedDescription's maps.
// order contains an array of nanoids that represent the order of the content blocks as they are in the original file.
func (ctx *RunContext) ParseSingleLanguageDescription(markdownRaw string) (title HTMLString, blocks []ContentBlock, footnotes Footnotes, abbreviations Abbreviations) {
	markdownRaw = HandleAltMediaEmbedSyntax(markdownRaw)
	htmlRaw := MarkdownToHTML(markdownRaw)
	htmlTree := soup.HTMLParse(htmlRaw)
	blocks = make([]ContentBlock, 0)
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
			block := ContentBlock{
				Type:   "media",
				Anchor: slugify.Marshal(firstChild.Attrs()["src"]),
				Media: Media{
					Alt:            alt,
					Caption:        firstChild.Attrs()["title"],
					RelativeSource: FilePathInsidePortfolioFolder(firstChild.Attrs()["src"]),
					Attributes:     attributes,
				},
			}
			block.ID = block.generateID()
			blocks = append(blocks, block)
		} else if childrenCount == 1 && firstChild.NodeValue == "a" {
			// An isolated link
			block := ContentBlock{
				Type:   "link",
				Anchor: slugify.Marshal(firstChild.FullText(), true),
				Link: Link{
					Text:  innerHTML(firstChild),
					Title: firstChild.Attrs()["title"],
					URL:   firstChild.Attrs()["href"],
				},
			}
			block.ID = block.generateID()
			blocks = append(blocks, block)
		} else if regexpMatches(PatternAbbreviationDefinition, string(innerHTML(paragraph))) {
			// An abbreviation definition
			groups := regexpGroups(PatternAbbreviationDefinition, string(innerHTML(paragraph)))
			abbreviations[groups[1]] = groups[2]
		} else if regexpMatches(PatternLanguageMarker, string(innerHTML(paragraph))) {
			// A language marker (ignored)
			continue
		} else {
			// A paragraph (anything else)
			block := ContentBlock{
				Type:   "paragraph",
				Anchor: paragraph.Attrs()["id"],
				Paragraph: Paragraph{
					Content: HTMLString(paragraph.HTML()),
				},
			}
			block.ID = block.generateID()
			blocks = append(blocks, block)
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
	for i, block := range blocks {
		if block.Type != "paragraph" {
			continue
		}
		if strings.HasPrefix(string(block.Paragraph.Content), "<pre>") && strings.HasSuffix(string(block.Paragraph.Content), "</pre>") {
			// Dont insert <abbr>s while in <pre> text
			continue
		}
		blocks[i].Paragraph = ReplaceAbbreviations(block.Paragraph, abbreviations)
	}

	LogDebug("Parsed description into blocks: %#v", blocks)
	return
}

// trimHTMLWhitespace removes whitespace from the beginning and end of an HTML string, also removing leading & trailing <br> tags.
func trimHTMLWhitespace(rawHTML HTMLString) HTMLString {
	rawHTML = HTMLString(strings.TrimSpace(string(rawHTML)))
	for _, toRemove := range []string{"<br>", "<br />", "<br/>"} {
		for strings.HasPrefix(string(rawHTML), toRemove) {
			rawHTML = HTMLString(strings.TrimPrefix(string(rawHTML), toRemove))
		}
		for strings.HasSuffix(string(rawHTML), toRemove) {
			rawHTML = HTMLString(strings.TrimSuffix(string(rawHTML), toRemove))
		}
	}
	return rawHTML
}

// HandleAltMediaEmbedSyntax handles the >[...](...) syntax by replacing it in htmlRaw with ![...](...).
func HandleAltMediaEmbedSyntax(markdownRaw string) string {
	pattern := regexp.MustCompile(`(?m)^>(\[[^\]]+\]\([^)]+\)\s*)$`)
	return pattern.ReplaceAllString(markdownRaw, "!$1")
}

// ExtractAttributesFromAlt extracts sigils from the end of the alt atetribute, returns the alt without them as well as the parse result.
func ExtractAttributesFromAlt(alt string) (string, MediaAttributes) {
	attrs := MediaAttributes{
		Controls: true, // Controls is added by default, others aren't
	}
	lastRune, _ := utf8.DecodeLastRuneInString(alt)
	// If there are no attributes in the alt string, the first (last in the alt string) will not be an attribute character.
	if !isMediaEmbedAttribute(lastRune) {
		return alt, attrs
	}
	altText := ""
	// We iterate backwardse:
	// if there are attributes, they'll be at the end of the alt text separated by a space
	inAttributesZone := true
	for i := len([]rune(alt)) - 1; i >= 0; i-- {
		char := []rune(alt)[i]
		if char == ' ' && inAttributesZone {
			inAttributesZone = false
			continue
		}
		if inAttributesZone {
			if char == RuneAutoplay {
				attrs.Autoplay = true
				attrs.Muted = true
			} else if char == RuneLoop {
				attrs.Loop = true
			} else if char == RuneHideControls {
				attrs.Controls = false
				attrs.Playsinline = true
			}
		} else {
			altText = string(char) + altText
		}
	}
	return altText, attrs
}

func isMediaEmbedAttribute(char rune) bool {
	return char == RuneAutoplay || char == RuneLoop || char == RuneHideControls
}

// innerHTML returns the HTML string of what's _inside_ the given element, just like JS' `element.innerHTML`.
func innerHTML(element soup.Root) HTMLString {
	var innerHTML string
	for _, child := range element.Children() {
		innerHTML += child.HTML()
	}
	if innerHTML == "" {
		innerHTML = element.HTML()
	}
	return HTMLString(innerHTML)
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
		processed = HTMLString(replacePattern.ReplaceAllString(string(paragraph.Content), "<abbr title=\""+definition+"\">"+name+"</abbr>"))
	}

	return Paragraph{Content: processed}
}
