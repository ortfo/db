package ortfodb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

func (ctx *RunContext) CreateDescriptionFile(workId string, metadataItems []string) error {
	output := ""
	metadata := WorkMetadata{
		Private: true,
	}
	outputPath := filepath.Join(ctx.PathToWorkFolder(workId), ctx.Config.ScatteredModeFolder, "description.md")

	// Construct the work metadata
	for _, item := range metadataItems {
		err := decodeMetadataItem(item, &metadata)
		if err != nil {
			return fmt.Errorf("while decoding metadata item %q: %w", item, err)
		}

	}
	if metadata.Started == "" {
		detectedStartDate, err := DetectStartDate(ctx.PathToWorkFolder(workId))
		if err != nil {
			ctx.LogWarning("while detecting start date of %s: %s", workId, err)
		} else {
			metadata.Started = detectedStartDate.Format("2006-01-02")
			LogCustom("Detected", "cyan", "start date to be %s", metadata.Started)
		}
	}

	output += "---\n"
	marshaledMetadata, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("while marshaling metadata of %s to yaml: %w", workId, err)
	}

	output += string(marshaledMetadata)
	output += "---\n\n"

	output += "# " + titleCase(workId) + "\n\n"

	if _, err := os.Stat(ctx.PathToWorkFolder(workId)); os.IsNotExist(err) {
		ctx.LogError("folder for given work %s does not exist.", workId)
		return nil
	}
	os.MkdirAll(filepath.Dir(outputPath), 0o755)
	os.WriteFile(outputPath, []byte(output), 0o644)
	LogCustom("Created", "green", "description.md file at [bold]%s[reset]", outputPath)
	return nil
}
