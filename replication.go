package ortfodb

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/anaskhan96/soup"
	"github.com/davecgh/go-spew/spew"
	"github.com/docopt/docopt-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// RunCommandReplicate runs the command 'replicate' given parsed CLI args from docopt.
func RunCommandReplicate(args docopt.Opts) error {
	// TODO: validate database.json
	var parsedDatabase []AnalyzedWork
	json := jsoniter.ConfigFastest
	setJSONNamingStrategy(lowerCaseWithUnderscores)
	databaseFilepath, err := args.String("<from-filepath>")
	if err != nil {
		return err
	}
	targetDatabasePath, err := args.String("<to-directory>")
	if err != nil {
		return err
	}
	content, err := readFileBytes(databaseFilepath)
	if err != nil {
		return err
	}
	validated, validationErrors, err := validateWithJSONSchema(string(content), databaseJSONSchema)
	if err != nil {
		return err
	}
	if !validated {
		DisplayValidationErrors(validationErrors, "database JSON")
		return nil
	}
	err = json.Unmarshal(content, &parsedDatabase)
	if err != nil {
		return err
	}
	ctx := RunContext{
		Config: &Configuration{},
	}
	defer fmt.Print("\033[2K\r\n")
	err = ReplicateAll(ctx, targetDatabasePath, parsedDatabase)
	if err != nil {
		return err
	}
	return nil
}

// ReplicateAll recreates a database inside targetDatabase containing all the works in works.
func ReplicateAll(ctx RunContext, targetDatabase string, works []AnalyzedWork) error {
	for _, work := range works {
		ctx.mu.Lock()
		ctx.CurrentWorkID = work.ID
		ctx.mu.Unlock()
		// ctx.Status() TODO
		err := ReplicateOne(targetDatabase, work)
		if err != nil {
			return err
		}
		ctx.Progress.Current++
	}
	return nil
}

// ReplicateOne creates a description.md file in targetDatabase in the correct folder in order to replicate Work.
func ReplicateOne(targetDatabase string, work AnalyzedWork) error {
	//TODO: make file mode configurable
	workDirectory := path.Join(targetDatabase, work.ID)
	os.MkdirAll(workDirectory, os.FileMode(0o0666))
	file, err := os.Create(path.Join(workDirectory, "description.md"))
	if err != nil {
		return err
	}
	defer file.Close()
	description, err := ReplicateDescription(work)
	if err != nil {
		return err
	}
	_, err = file.WriteString(description)
	if err != nil {
		return err
	}
	return nil
}

// ReplicateDescription reconstructs the contents of a description.md file from a Work struct.
func ReplicateDescription(work AnalyzedWork) (string, error) {
	var result string
	// Start with the YAML header, this one is never localized
	yamlHeader, err := replicateMetadata(work.Metadata)
	if err != nil {
		return "", err
	}
	result += yamlHeader + "\n"
	// TODO get rid of "default" language behavior
	// if a file has NO language markers, auto-insert ":: (machine's language)" before parsing.
	for language := range work.Content {
		result += replicateLanguageMarker(language) + "\n\n"
		replicatedBlock, err := replicateLocalizedBlock(work, language)
		if err != nil {
			return "", err
		}
		result += replicatedBlock + "\n\n"
	}
	return strings.TrimSpace(result), nil
}

func replicateLocalizedBlock(work AnalyzedWork, language string) (string, error) {
	var result string
	end := "\n\n"
	content := work.Content[language]
	// Abbreviations will be stored here to declare them in the markdown
	abbreviations := make(Abbreviations)
	// Start with the title
	if content.Title != "" {
		result += replicateTitle(content.Title) + end
	}
	// Then, for each block (ordered by the layout)
	spew.Dump(work)
	for _, block := range content.Blocks {
		fmt.Printf("replicating %s block #%s", block.Type, block.ID)
		switch block.Type {
		case "media":
			result += replicateMediaEmbed(block.Media) + end
		case "link":
			result += replicateLink(block.Link) + end
		case "paragraph":
			replicatedParagraph, err := replicateParagraph(block.Paragraph)
			if err != nil {
				return "", err
			}
			// This is not finished: we need to properly translate to markdown abbreviations & footnotes
			parsedHTML := soup.HTMLParse(string(block.Content))
			abbreviations = merge(abbreviations, collectAbbreviations(parsedHTML))
			replicatedParagraph = transformAbbreviations(parsedHTML, replicatedParagraph)
			replicatedParagraph = transformFootnoteReferences(replicatedParagraph)
			result += replicatedParagraph + end
		default: // nothing
		}
	}
	for name, content := range content.Footnotes {
		result += replicateFootnoteDefinition(name, string(content)) + end
	}
	result += replicateAbbreviations(abbreviations)
	return result, nil
}

func replicateLanguageMarker(language string) string {
	return ":: " + language
}

// transformFootnoteReferences turns HTML references to footnotes into markdown ones.
func transformFootnoteReferences(markdown string) string {
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
func transformAbbreviations(htmlSoup soup.Root, markdown string) string {
	transformedMarkdown := markdown
	for _, abbr := range htmlSoup.FindAll("abbr") {
		transformedMarkdown = strings.ReplaceAll(transformedMarkdown, abbr.HTML(), abbr.FullText())
	}
	return transformedMarkdown
}

func collectAbbreviations(htmlSoup soup.Root) Abbreviations {
	abbreviations := make(Abbreviations)
	for _, abbr := range htmlSoup.FindAll("abbr") {
		abbreviations[abbr.FullText()] = abbr.Attrs()["title"]
	}
	return abbreviations
}

// We replicate all abbreviations in one function to avoid duplicates.
func replicateAbbreviations(abbreviations Abbreviations) string {
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

func replicateFootnoteDefinition(name string, content string) string {
	return "[^" + name + "]: " + content
}

func replicateLink(link Link) string {
	if link.Title != "" {
		return "[" + link.Text.String() + `](` + link.URL + ` "` + link.Title + `")`
	}
	return "[" + link.Text.String() + "](" + link.URL + ")"
}

func replicateTitle(title HTMLString) string {
	return "# " + title.Markdown()
}

func replicateMetadata(metadata WorkMetadata) (string, error) {
	metadataOut := make(map[string]interface{})
	mapstructure.Decode(metadata, &metadataOut)
	yamlBytes, err := yaml.Marshal(metadataOut)
	if err != nil {
		return "", err
	}
	return "---\n" + string(yamlBytes) + "---", nil
}

func replicateMediaAttributesString(attributes MediaAttributes) string {
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
func replicateMediaEmbed(media Media) string {
	if media.Title != "" {
		return fmt.Sprintf(`![%s %s](%s "%s")`, media.Alt, replicateMediaAttributesString(media.Attributes), string(media.RelativeSource), media.Title)
	}
	return fmt.Sprintf(`![%s %s](%s)`, media.Alt, replicateMediaAttributesString(media.Attributes), string(media.RelativeSource))
}

func replicateParagraph(p Paragraph) (string, error) {
	markdown := p.Content.Markdown()
	if strings.TrimSpace(markdown) == "" {
		markdown = "<p></p>"
	}
	var result string
	if p.Anchor != "" {
		result = "{#" + p.Anchor + "}\n" + markdown
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
