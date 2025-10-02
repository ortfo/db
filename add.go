package ortfodb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/charmbracelet/huh"
	ll "github.com/gwennlbh/label-logger-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gopkg.in/yaml.v3"
)

func gitCommitDate(workingDirectory string, gitFlags ...string) (time.Time, error) {
	// Shell out to git
	gitLog := exec.Command("git", "log", "--format=%aI")
	gitLog.Args = append(gitLog.Args, gitFlags...)
	gitLog.Dir = workingDirectory
	out, err := gitLog.Output()
	if err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			return time.Now(), fmt.Errorf("while getting creation date from git commits in %s: %s (%w)", workingDirectory, string(err.Stderr), err)
		default:
			return time.Now(), fmt.Errorf("while getting creation date from git commits in %s: %w", workingDirectory, err)
		}
	}

	// get first line
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return time.Now(), fmt.Errorf("while getting creation date from git commits in %s: git returned no output", workingDirectory)
	}
	return time.Parse(time.RFC3339, lines[0])
}

func gitRemoteURL(workingDirectory string) (string, error) {
	// Shell out to git
	gitRemote := exec.Command("git", "remote", "get-url", "origin")
	gitRemote.Dir = workingDirectory
	out, err := gitRemote.Output()
	if err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			return "", fmt.Errorf("while getting remote URL from git in %s: %s (%w)", workingDirectory, string(err.Stderr), err)
		default:
			return "", fmt.Errorf("while getting remote URL from git in %s: %w", workingDirectory, err)
		}
	}

	return strings.TrimSpace(string(out)), nil
}

func FirstGitCommitDate(workingDirectory string) (time.Time, error) {
	return gitCommitDate(workingDirectory, "--reverse")
}

func LastGitCommitDate(workingDirectory string) (time.Time, error) {
	return gitCommitDate(workingDirectory)
}

// titleCase replaces underscores and dashes with spaces and capitalizes the first letter of each word.
func titleCase(s string) string {
	return cases.Title(language.English).String(
		regexp.MustCompile(`[-_]`).ReplaceAllString(strings.TrimSpace(s), " "),
	)
}

func isGitRepo(workingDirectory string) bool {
	_, err := os.Stat(filepath.Join(workingDirectory, ".git"))
	return err == nil
}

func DetectStartDate(workingDirectory string) (time.Time, error) {
	// If in a git repo, get the date of the first commit
	if isGitRepo(workingDirectory) {
		return FirstGitCommitDate(workingDirectory)
	}
	return time.Now(), fmt.Errorf("no way to autodetect start date of %s", workingDirectory)
}

func decodeMetadataItem(item string, metadata *WorkMetadata) error {
	parts := strings.SplitN(item, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid metadata item: %s", item)
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	switch key {
	case "started":
		metadata.Started = value
	case "finished":
		metadata.Finished = value
	case "tag":
	case "tags":
		metadata.Tags = append(metadata.Tags, value)
	case "madewith":
	case "using":
		metadata.MadeWith = append(metadata.MadeWith, value)
	default:
		return fmt.Errorf("unknown metadata key: %s", key)
	}
	return nil
}

// fromReadme extracts the title from the readme and returns the entire readme content
func fromReadme(readmePath string) (title string, firstParagraph string, err error) {
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		err = fmt.Errorf("while reading %s: %w", readmePath, err)
		return
	}

	html, err := MarkdownToHTML(string(readme))
	if err != nil {
		err = fmt.Errorf("while parsing markdown from README: %w", err)
		return
	}

	readmeTree := soup.HTMLParse(html)
	title = readmeTree.Find("h1").FullText()
	firstParagraph = HTMLString(readmeTree.Find("p").HTML()).Markdown()
	return
}

func (ctx *RunContext) CreateDescriptionFile(workId string, metadataItems []string, forceOverwrite bool) (string, error) {
	output := ""
	outputPath := filepath.Join(ctx.PathToWorkFolder(workId), ctx.Config.ScatteredModeFolder, "description.md")

	if _, err := os.Stat(ctx.PathToWorkFolder(workId)); os.IsNotExist(err) {
		return outputPath, fmt.Errorf("folder for given work %s (%s) does not exist", workId, ctx.PathToWorkFolder(workId))
	}

	if _, err := os.Stat(outputPath); err == nil && !forceOverwrite {
		confirmOverwrite := false
		huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("%s already exists", outputPath)).
				Description("Overwrite it?").Value(&confirmOverwrite),
		)).Run()
		if !confirmOverwrite {
			return outputPath, nil
		}
	}

	allTags, err := ctx.LoadTagsRepository()
	if err != nil {
		return outputPath, fmt.Errorf("while reading all available tags: %w", err)
	}

	allTagsOptions := make([]huh.Option[string], 0, len(allTags))
	for _, tag := range allTags {
		allTagsOptions = append(allTagsOptions, huh.NewOption(tag.String(), tag.String()))
	}

	allTechs, err := ctx.LoadTechnologiesRepository()
	if err != nil {
		return outputPath, fmt.Errorf("while reading all available technologies: %w", err)
	}

	allTechsOptions := make([]huh.Option[string], 0, len(allTechs))
	for _, tech := range allTechs {
		allTechsOptions = append(allTechsOptions, huh.NewOption(tech.String(), tech.Slug))
	}

	defaultProjectTitle := titleCase(workId)
	defaultSummary := ""

	readmePath := filepath.Join(ctx.PathToWorkFolder(workId), "README.md")
	if fileExists(readmePath) {
		readmeTitle, readmeBody, err := fromReadme(readmePath)
		if err != nil {
			ll.WarnDisplay("couldn't extract info from README.md", err)
		} else {
			if readmeTitle != "" {
				defaultProjectTitle = readmeTitle
			}
			defaultSummary = readmeBody
		}
	}

	detectedStartDate, err := DetectStartDate(ctx.PathToWorkFolder(workId))
	defaultStartedAt := ""
	if err != nil {
		ll.WarnDisplay("while detecting start date of %s", err, workId)
	} else {
		defaultStartedAt = detectedStartDate.Format("2006-01-02")
		ll.Log("Detected", "cyan", "start date to be [bold][blue]%s[reset]", defaultStartedAt)
	}

	startedAtPlaceholder := "YYYY-MM-DD"
	if defaultStartedAt != "" {
		startedAtPlaceholder = defaultStartedAt
	}

	metadata := WorkMetadata{
		Private:  true,
		Tags:     []string{},
		MadeWith: []string{},
	}

	autodetectedTechs, err := ctx.DetectTechnologies(workId)
	if err != nil {
		ll.Warn(ll.FormatErrors(fmt.Errorf("while autodetecting technologies for %s: %w", workId, err)))
	} else {
		displayTags := make([]string, 0, len(autodetectedTechs))
		for _, tech := range autodetectedTechs {
			metadata.MadeWith = append(metadata.MadeWith, tech.Slug)
			displayTags = append(displayTags, tech.String())
		}
		if len(metadata.MadeWith) > 0 {
			ll.Log("Detected", "cyan", "technologies to be %s", ll.List(displayTags, "[bold][blue]%s[reset]", ", "))
		}
	}

	autodetectedTags, err := ctx.DetectTags(workId, autodetectedTechs)
	if err != nil {
		ll.WarnDisplay("while autodetecting tags for %s", err, workId)
	} else {
		for _, tag := range autodetectedTags {
			metadata.Tags = append(metadata.Tags, tag.String())
		}
		if len(metadata.Tags) > 0 {
			ll.Log("Detected", "cyan", "tags to be %s", ll.List(metadata.Tags, "[bold][blue]%s[reset]", ", "))
		}
	}

	var projectTitle string
	var summary string

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Placeholder(defaultProjectTitle).Value(&projectTitle),

			huh.NewText().Title("Summary").Description("A short description of the work").Placeholder(defaultSummary).Value(&summary),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Technologies").
				Description("What was the work made with?").
				Filterable(true).
				Value(&metadata.MadeWith).
				Options(allTechsOptions...).
				Validate(func(s []string) error {
					ll.Debug("Selected %v", s)
					return nil
				}).
				Height(2+6),

			huh.NewMultiSelect[string]().
				Title("Tags").
				Description("Categorize your work").
				Filterable(true).
				Value(&metadata.Tags).
				Options(allTagsOptions...).
				Height(2+6),
		),
		huh.NewGroup(
			huh.NewInput().Description("When did you start working on this?").Placeholder(startedAtPlaceholder).Value(&metadata.Started),

			huh.NewConfirm().Title("Work in progress").Description("What's the status?").Value(&metadata.WIP).Affirmative("WIP").Negative("Finished"),
		),
	).Run()

	if err != nil {
		return outputPath, fmt.Errorf("while getting your answers: %w", err)
	}

	if !metadata.WIP {
		defaultFinishedAt := time.Now().Format("2006-01-02")
		if finishedAtFromGit, err := LastGitCommitDate(ctx.PathToWorkFolder(workId)); err == nil {
			defaultFinishedAt = finishedAtFromGit.Format("2006-01-02")
			ll.Log("Detected", "cyan", "finish date to be [bold][blue]%s[reset]", defaultFinishedAt)
		}

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Description("When did you finish working on this?").Placeholder(defaultFinishedAt).Value(&metadata.Finished),
			),
		).Run()
		if err != nil {
			return outputPath, fmt.Errorf("while getting your answer: %w", err)
		}

		if metadata.Finished == "" {
			metadata.Finished = defaultFinishedAt
		}
	}

	if projectTitle == "" {
		projectTitle = defaultProjectTitle
	}

	if metadata.Started == "" {
		metadata.Started = defaultStartedAt
	}

	// Construct the work metadata
	for _, item := range metadataItems {
		err := decodeMetadataItem(item, &metadata)
		if err != nil {
			return outputPath, fmt.Errorf("while decoding metadata item %q: %w", item, err)
		}
	}

	output += "---\n"
	marshaledMetadata, err := yaml.Marshal(metadata)
	if err != nil {
		return outputPath, fmt.Errorf("while marshaling metadata of %s to yaml: %w", workId, err)
	}

	output += string(marshaledMetadata)
	output += "---\n\n"

	output += "# " + projectTitle + "\n\n"
	output += summary + "\n\n"

	if isGitRepo(ctx.PathToWorkFolder(workId)) {
		remoteURL, err := gitRemoteURL(ctx.PathToWorkFolder(workId))
		if err == nil {
			output += fmt.Sprintf("[Source code](%s)\n\n", remoteURL)
		}
	}

	os.MkdirAll(filepath.Dir(outputPath), 0o755)
	os.WriteFile(outputPath, []byte(output), 0o644)
	ll.Log("Created", "green", "description.md file at [bold]%s[reset]", outputPath)
	return outputPath, nil
}
