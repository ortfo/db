package ortfodb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gopkg.in/yaml.v3"
)

func FirstGitCommitDate(workingDirectory string) (time.Time, error) {
	// Shell out to git
	gitLog := exec.Command("git", "log", "--reverse", "--format=%aI")
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

// titleCase replaces underscores and dashes with spaces and capitalizes the first letter of each word.
func titleCase(s string) string {
	return cases.Title(language.English).String(
		regexp.MustCompile(`[-_]`).ReplaceAllString(strings.TrimSpace(s), " "),
	)
}

func DetectStartDate(workingDirectory string) (time.Time, error) {
	// If in a git repo, get the date of the first commit
	if _, err := os.Stat(filepath.Join(workingDirectory, ".git")); err == nil {
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

func (ctx *RunContext) CreateDescriptionFile(workId string, metadataItems []string, forceOverwrite bool) (string, error) {
	output := ""
	outputPath := filepath.Join(ctx.PathToWorkFolder(workId), ctx.Config.ScatteredModeFolder, "description.md")

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

	detectedStartDate, err := DetectStartDate(ctx.PathToWorkFolder(workId))
	defaultStartedAt := ""
	if err != nil {
		ctx.LogWarning(formatErrors(fmt.Errorf("while detecting start date of %s: %w", workId, err)))
	} else {
		defaultStartedAt = detectedStartDate.Format("2006-01-02")
		LogCustom("Detected", "cyan", "start date to be [bold][blue]%s[reset]", defaultStartedAt)
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
		ctx.LogWarning(formatErrors(fmt.Errorf("while autodetecting technologies for %s: %w", workId, err)))
	} else {
		displayTags := make([]string, 0, len(autodetectedTechs))
		for _, tech := range autodetectedTechs {
			metadata.MadeWith = append(metadata.MadeWith, tech.Slug)
			displayTags = append(displayTags, tech.String())
		}
		if len(metadata.MadeWith) > 0 {
			LogCustom("Detected", "cyan", "technologies to be %s", formatList(displayTags, "[bold][blue]%s[reset]", ", "))
		}
	}

	autodetectedTags, err := ctx.DetectTags(workId, autodetectedTechs)
	if err != nil {
		ctx.LogWarning("while autodetecting tags for %s: %s", workId, err)
	} else {
		for _, tag := range autodetectedTags {
			metadata.Tags = append(metadata.Tags, tag.String())
		}
		if len(metadata.Tags) > 0 {
			LogCustom("Detected", "cyan", "tags to be %s", formatList(metadata.Tags, "[bold][blue]%s[reset]", ", "))
		}
	}

	var projectTitle string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Description("Title of the work").Placeholder(defaultProjectTitle).Value(&projectTitle),

			huh.NewMultiSelect[string]().
				Title("Technologies").
				Description("What was the work made with?").
				Filterable(true).
				Value(&metadata.MadeWith).
				Options(allTechsOptions...).
				Validate(func(s []string) error {
					ctx.LogDebug("Selected %v", s)
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

			huh.NewInput().Description("When did you start working on this?").Placeholder(startedAtPlaceholder).Value(&metadata.Started),

			huh.NewConfirm().Title("Work in progress").Description("What's the status?").Value(&metadata.WIP).Affirmative("WIP").Negative("Finished"),
		),
	)

	err = form.Run()
	if err != nil {
		return outputPath, fmt.Errorf("while getting your answers: %w", err)
	}

	if !metadata.WIP {
		defaultFinishedAt := time.Now().Format("2006-01-02")

		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Description("When did you finish working on this?").Placeholder(defaultFinishedAt).Value(&metadata.Finished),
			),
		)

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

	if _, err := os.Stat(ctx.PathToWorkFolder(workId)); os.IsNotExist(err) {
		ctx.LogError("folder for given work %s does not exist.", workId)
		return outputPath, nil
	}
	os.MkdirAll(filepath.Dir(outputPath), 0o755)
	os.WriteFile(outputPath, []byte(output), 0o644)
	LogCustom("Created", "green", "description.md file at [bold]%s[reset]", outputPath)
	return outputPath, nil
}
