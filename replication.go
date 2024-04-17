package ortfodb

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/anaskhan96/soup"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// ReplicateAll recreates a database inside targetDatabase containing all the works in works.
func (ctx *RunContext) ReplicateAll(targetDatabase string, works Database) error {
	for _, work := range works {
		err := ctx.ReplicateOne(targetDatabase, work)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReplicateOne creates a description.md file in targetDatabase in the correct folder in order to replicate Work.
func (ctx *RunContext) ReplicateOne(targetDatabase string, work AnalyzedWork) error {
	//TODO: make file mode configurable
	workDirectory := path.Join(targetDatabase, work.ID)
	os.MkdirAll(workDirectory, os.FileMode(0o0777))
	description, err := ctx.ReplicateDescription(work)
	if err != nil {
		return fmt.Errorf("while replicating %s: %w", work.ID, err)
	}

	os.WriteFile(path.Join(workDirectory, "description.md"), []byte(description), os.FileMode(0o0777))
	return nil
}

// ReplicateDescription reconstructs the contents of a description.md file from a Work struct.
func (ctx *RunContext) ReplicateDescription(work AnalyzedWork) (string, error) {
	var result string
	// Start with the YAML header, this one is never localized
	yamlHeader, err := ctx.replicateMetadata(work.Metadata)
	if err != nil {
		return "", err
	}
	result += yamlHeader + "\n"
	// TODO get rid of "default" language behavior
	// if a file has NO language markers, auto-insert ":: (machine's language)" before parsing.
	for language := range work.Content {
		result += ctx.replicateLanguageMarker(language) + "\n\n"
		replicatedBlock, err := ctx.replicateLocalizedBlock(work, language)
		if err != nil {
			return "", err
		}
		result += replicatedBlock + "\n\n"
	}
	return strings.TrimSpace(result), nil
}

func (ctx *RunContext) replicateLocalizedBlock(work AnalyzedWork, language string) (string, error) {
	var result string
	end := "\n\n"
	content := work.Content[language]
	// Abbreviations will be stored here to declare them in the markdown
	abbreviations := make(Abbreviations)
	// Start with the title
	if content.Title != "" {
		result += ctx.replicateTitle(content.Title) + end
	}
	// Then, for each block (ordered by the layout)
	// spew.Dump(work)
	for _, block := range content.Blocks {
		LogDebug("replicating %s block #%s", block.Type, block.ID)
		switch block.Type {
		case "media":
			result += ctx.replicateMediaEmbed(block.Media) + end
		case "link":
			result += ctx.replicateLink(block.Link) + end
		case "paragraph":
			replicatedParagraph, err := ctx.replicateParagraph(block.Anchor, block.Paragraph)
			if err != nil {
				return "", err
			}
			// This is not finished: we need to properly translate to markdown abbreviations & footnotes
			parsedHTML := soup.HTMLParse(string(block.Content))
			abbreviations = merge(abbreviations, ctx.collectAbbreviations(parsedHTML))
			replicatedParagraph = ctx.transformAbbreviations(parsedHTML, replicatedParagraph)
			replicatedParagraph = ctx.transformFootnoteReferences(replicatedParagraph)
			result += replicatedParagraph + end
		default: // nothing
		}
	}
	for name, content := range content.Footnotes {
		result += ctx.replicateFootnoteDefinition(name, string(content)) + end
	}
	result += ctx.replicateAbbreviations(abbreviations)
	return result, nil
}

func (ctx *RunContext) replicateLanguageMarker(language string) string {
	return ":: " + language
}

// transformFootnoteReferences turns HTML references to footnotes into markdown ones.
func (ctx *RunContext) transformFootnoteReferences(markdown string) string {
	pattern := regexp.MustCompile(`\[(\d+)\]\(#fn:([^)]+)\)`)
	lines := strings.Split(markdown, "\n")
	transformedMarkdown := markdown
	for _, line := range lines {
		if pattern.MatchString(line) {
			for _, groups := range pattern.FindAllStringSubmatch(line, -1) {
				transformedMarkdown = strings.ReplaceAll(transformedMarkdown, groups[0], "[^"+groups[2]+"]")
			}
		}
	}
	return transformedMarkdown
}

// Remove markup from abbreviations.
func (ctx *RunContext) transformAbbreviations(htmlSoup soup.Root, markdown string) string {
	transformedMarkdown := markdown
	for _, abbr := range htmlSoup.FindAll("abbr") {
		transformedMarkdown = strings.ReplaceAll(transformedMarkdown, abbr.HTML(), abbr.FullText())
	}
	return transformedMarkdown
}

func (ctx *RunContext) collectAbbreviations(htmlSoup soup.Root) Abbreviations {
	abbreviations := make(Abbreviations)
	for _, abbr := range htmlSoup.FindAll("abbr") {
		abbreviations[abbr.FullText()] = abbr.Attrs()["title"]
	}
	return abbreviations
}

// We replicate all abbreviations in one function to avoid duplicates.
func (ctx *RunContext) replicateAbbreviations(abbreviations Abbreviations) string {
	var result string
	// Stores all the alread-replicated abbreviations' names (to handle duplicates)
	replicated := make([]string, 0, len(abbreviations))
	for name, definition := range abbreviations {
		if stringInSlice(replicated, name) {
			continue
		}
		result += "*[" + name + "]: " + definition + "\n"
		replicated = append(replicated, definition)
	}
	return result
}

func (ctx *RunContext) replicateFootnoteDefinition(name string, content string) string {
	return "[^" + name + "]: " + content
}

func (ctx *RunContext) replicateLink(link Link) string {
	if link.Title != "" {
		return "[" + link.Text.String() + `](` + link.URL + ` "` + link.Title + `")`
	}
	return "[" + link.Text.String() + "](" + link.URL + ")"
}

func (ctx *RunContext) replicateTitle(title HTMLString) string {
	return "# " + title.Markdown()
}

func (ctx *RunContext) replicateMetadata(metadata WorkMetadata) (string, error) {
	metadataOut := make(map[string]interface{})
	mapstructure.Decode(metadata, &metadataOut)
	yamlBytes, err := yaml.Marshal(metadataOut)
	if err != nil {
		return "", err
	}
	return "---\n" + string(yamlBytes) + "---", nil
}

func (ctx *RunContext) replicateMediaAttributesString(attributes MediaAttributes) string {
	result := ""
	if attributes.Autoplay {
		result += string(RuneAutoplay)
	}
	if !attributes.Controls {
		result += string(RuneHideControls)
	}
	if attributes.Loop {
		result += string(RuneLoop)
	}
	return result
}

// TODO: configure whether to use >[]() syntax: never, or only for non-images
func (ctx *RunContext) replicateMediaEmbed(media Media) string {
	if media.Caption != "" {
		return fmt.Sprintf(`![%s %s](%s "%s")`, media.Alt, ctx.replicateMediaAttributesString(media.Attributes), string(media.RelativeSource), media.Caption)
	}
	return fmt.Sprintf(`![%s %s](%s)`, media.Alt, ctx.replicateMediaAttributesString(media.Attributes), string(media.RelativeSource))
}

func (ctx *RunContext) replicateParagraph(anchor string, p Paragraph) (string, error) {
	markdown := p.Content.Markdown()
	if strings.TrimSpace(markdown) == "" {
		markdown = "<p></p>"
	}
	var result string
	if anchor != "" {
		result = "{#" + anchor + "}\n" + markdown
	} else {
		result = markdown
	}
	return result, nil
}

func (html HTMLString) Markdown() string {
	// TODO: configurable domain for translating relative to absolute URLS from ortfodb.yaml
	converter := html2md.NewConverter("", true, nil)
	result, err := converter.ConvertString(string(html))
	if err != nil {
		return html.String()
	}
	return result
}
