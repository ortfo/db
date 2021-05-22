package ortfodb

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/anaskhan96/soup"
	"github.com/docopt/docopt-go"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
)

// RunCommandReplicate runs the command 'replicate' given parsed CLI args from docopt
func RunCommandReplicate(args docopt.Opts) error {
	// TODO: validate database.json
	var parsedDatabase []Work
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
		Progress: struct {
			Current int
			Total   int
		}{Total: len(parsedDatabase)},
	}
	defer fmt.Print("\033[2K\r\n")
	err = ReplicateAll(ctx, targetDatabasePath, parsedDatabase)
	if err != nil {
		return err
	}
	return nil
}

// ReplicateAll recreates a database inside targetDatabase containing all the works in `works`
func ReplicateAll(ctx RunContext, targetDatabase string, works []Work) error {
	for _, work := range works {
		ctx.CurrentProject = work.ID
		ctx.Status("Replicating")
		err := ReplicateOne(targetDatabase, work)
		if err != nil {
			return err
		}
		ctx.Progress.Current++
	}
	return nil
}

// ReplicateOne creates a description.md file in targetDatabase in the correct folder in order to replicate Work
func ReplicateOne(targetDatabase string, work Work) error {
	//TODO: make file mode configurable
	workDirectory := path.Join(targetDatabase, work.ID)
	os.MkdirAll(workDirectory, os.FileMode(0o0666))
	file, err := os.Create(path.Join(workDirectory, "description.md"))
	if err != nil {
		return err
	}
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

// ReplicateDescription reconstructs the contents of a description.md file from a Work struct
func ReplicateDescription(work Work) (string, error) {
	var result string
	// Start with the YAML header, this one is never localized
	yamlHeader, err := replicateMetadata(work.Metadata)
	if err != nil {
		return "", err
	}
	result += yamlHeader + "\n"
	// Then, all the unlocalized stuff (language "default")
	replicatedBlock, err := replicateLocalizedBlock(work, "default")
	if err != nil {
		return "", err
	}
	result += replicatedBlock
	for _, language := range mapKeys(work.Title) {
		result += replicateLanguageMarker(language) + "\n\n"
		replicatedBlock, err := replicateLocalizedBlock(work, language)
		if err != nil {
			return "", err
		}
		result += replicatedBlock + "\n\n"
	}
	return strings.TrimSpace(result), nil
}

func replicateLocalizedBlock(work Work, language string) (string, error) {
	var result string
	end := "\n\n"
	// Abbreviations will be stored here to declare them in the markdown
	abbreviations := make([]Abbreviation, 0)
	// Start with the title
	if work.Title[language] != "" {
		result += replicateTitle(work.Title[language]) + end
	}
	// Then, every media
	for _, media := range work.Media[language] {
		result += replicateMediaEmbed(media) + end
	}
	for _, paragraph := range work.Paragraphs[language] {
		replicatedParagraph, err := replicateParagraph(paragraph)
		if err != nil {
			return "", err
		}
		// This is not finished: we need to properly translate to markdown abbreviations & footnotes
		parsedHTML := soup.HTMLParse(replicatedParagraph)
		abbreviations = append(abbreviations, collectAbbreviations(parsedHTML)...)
		replicatedParagraph = transformAbbreviations(parsedHTML, replicatedParagraph)
		replicatedParagraph = transformFootnoteReferences(replicatedParagraph)
		result += replicatedParagraph + end
	}
	for _, link := range work.Links[language] {
		result += replicateLink(link) + end
	}
	for _, footnote := range work.Footnotes[language] {
		result += replicateFootnoteDefinition(footnote) + end
	}
	result += replicateAbbreviations(abbreviations)
	return result, nil
}

func replicateLanguageMarker(language string) string {
	return ":: " + language
}

// transformFootnoteReferences turns HTML references to footnotes into markdown ones
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

// Remove markup from abbreviations
func transformAbbreviations(htmlSoup soup.Root, markdown string) string {
	transformedMarkdown := markdown
	for _, abbr := range htmlSoup.FindAll("abbr") {
		transformedMarkdown = strings.ReplaceAll(transformedMarkdown, abbr.HTML(), abbr.FullText())
	}
	return transformedMarkdown
}

func collectAbbreviations(htmlSoup soup.Root) []Abbreviation {
	abbreviations := make([]Abbreviation, 0)
	for _, abbr := range htmlSoup.FindAll("abbr") {
		abbreviations = append(abbreviations, Abbreviation{
			Definition: abbr.Attrs()["title"],
			Name:       abbr.FullText(),
		})
	}
	return abbreviations
}

// We replicate all abbreviations in one function to avoid duplicates
func replicateAbbreviations(abbreviations []Abbreviation) string {
	var result string
	// Stores all the alread-replicated abbreviations' names (to handle duplicates)
	replicated := make([]string, 0, len(abbreviations))
	for _, abbreviation := range abbreviations {
		if stringInSlice(replicated, abbreviation.Name) {
			continue
		}
		result += "*[" + abbreviation.Name + "]: " + abbreviation.Definition
		replicated = append(replicated, abbreviation.Name)
	}
	return result
}

func replicateFootnoteDefinition(footnote Footnote) string {
	return "[^" + footnote.Name + "]: " + footnote.Content
}

func replicateLink(link Link) string {
	if link.Title != "" {
		return "[" + link.Name + ` "` + link.Title + `"](` + link.URL + ")"
	}
	return "[" + link.Name + "](" + link.URL + ")"
}

func replicateTitle(title string) string {
	return "# " + title
}

func replicateMetadata(metadata map[string]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return "---\n" + string(yamlBytes) + "\n---", nil
}

//TODO: configure whether to use >[]() syntax: never, or only for non-images
func replicateMediaEmbed(media Media) string {
	if media.Title != "" {
		return "![" + media.Alt + ` "` + media.Title + `"](` + media.Source + ")"
	}
	return "![" + media.Alt + "](" + media.Source + ")"
}

func replicateParagraph(p Paragraph) (string, error) {
	markdown, err := htmlToMarkdown(p.Content)
	if err != nil {
		return "", err
	}
	var result string
	if p.ID != "" {
		result = "{#" + p.ID + "}\n" + markdown
	} else {
		result = markdown
	}
	return result, nil
}

func htmlToMarkdown(html string) (string, error) {
	// TODO: configurable domain for translating relative to absolute URLS from .portfoliodb.yml
	converter := html2md.NewConverter("", true, nil)
	return converter.ConvertString(html)
}
