package ortfodb

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"gopkg.in/yaml.v2"
)

type autodetectData struct {
	// Name is only used for debug purposes
	Name string

	Files             []string
	ContentConditions []string
}

type Tag struct {
	Singular         string   `yaml:"singular"`
	Plural           string   `yaml:"plural"`
	Description      string   `yaml:"description"`
	LearnMoreAt      string   `yaml:"learn more at"`
	Aliases          []string `yaml:"aliases"`
	DetectConditions struct {
		Files    []string `yaml:"files"`
		Search   []string `yaml:"search"`
		MadeWith []string `yaml:"made with"`
	} `yaml:"detect"`
}

func (t Tag) String() string {
	return t.Singular
}

func (t Tag) Detect(ctx *RunContext, workId string, techs []Technology) (bool, error) {
	for _, tech := range t.DetectConditions.MadeWith {
		for _, candidate := range techs {
			if candidate.ReferredToBy(tech) {
				return true, nil
			}
		}
	}
	return autodetectData{
		Name:              t.Singular,
		ContentConditions: t.DetectConditions.Search,
		Files:             t.DetectConditions.Files,
	}.Detect(ctx, workId)
}

type Technology struct {
	Slug        string `yaml:"slug"`
	Name        string `yaml:"name"`
	By          string `yaml:"by"`
	Description string `yaml:"description"`
	LearnMoreAt string `yaml:"learn more at"`

	Aliases []string `yaml:"aliases"`

	// Files contains a list of gitignore-style patterns. If the work contains any of the patterns specified, we consider that technology to be used in the work.
	Files []string `yaml:"files"`
	// Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a free-form unquoted string and PATH is a filepath relative to the work folder.
	// If CONTENT is found in PATH, we consider that technology to be used in the work.
	Autodetect []string `yaml:"autodetect"`
}

func (t Technology) String() string {
	return t.Name
}

func (t Technology) Detect(ctx *RunContext, workId string) (bool, error) {
	return autodetectData{
		Name:              t.Slug,
		ContentConditions: t.Autodetect,
		Files:             t.Files,
	}.Detect(ctx, workId)
}

// Detect returns true if this technology is detected as used in the work.
func (t autodetectData) Detect(ctx *RunContext, workId string) (matched bool, err error) {
	// Match files
	contentDetectionConditions := make(map[string][]string)
	contentDetectionFiles := make([]string, 0)
	for _, f := range t.ContentConditions {
		parts := strings.Split(f, " in ")
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid autodetect expression: %s", f)
		}
		content := parts[0]
		path := parts[1]
		contentDetectionFiles = append(contentDetectionFiles, path)

		if _, ok := contentDetectionConditions[path]; !ok {
			contentDetectionConditions[path] = []string{content}
		} else {
			contentDetectionConditions[path] = append(contentDetectionConditions[path], content)
		}
	}
	LogDebug("Starting auto-detect for %s: contentDetection map is %v", t, contentDetectionConditions)
	for _, f := range append(t.Files, contentDetectionFiles...) {
		_, isContentDetection := contentDetectionConditions[f]
		LogDebug("Auto-detecting %s in %s: %q: isContentDetection=%v", t, workId, f, isContentDetection)
		pat := gitignore.ParsePattern(f, nil)
		// Walk all files of the work folder (excl. hidden files unfortunately)
		err = fs.WalkDir(os.DirFS(ctx.PathToWorkFolder(workId)), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if filepath.Base(path) == "node_modules" {
				return fs.SkipDir
			}
			if filepath.Base(path) == ".venv" {
				return fs.SkipDir
			}

			pathFragments := make([]string, 0)
			for _, fragment := range strings.Split(path, string(os.PathSeparator)) {
				if fragment != "" {
					pathFragments = append(pathFragments, fragment)
				}
			}

			if isContentDetection {
				if path != f {
					return nil
				}
				contents, err := readFile(filepath.Join(ctx.PathToWorkFolder(workId), path))
				if err != nil {
					return fmt.Errorf("while reading contents of %s to check whether it contains one of %#v: %w", path, contentDetectionConditions[path], err)
				}

				for _, contentCondition := range contentDetectionConditions[path] {
					if strings.Contains(contents, contentCondition) {
						LogDebug("Auto-detected %s in %s: condition %q in %q met", t, workId, contentCondition, path)
						matched = true
						return filepath.SkipAll
					}
				}

			} else {
				result := pat.Match(pathFragments, d.IsDir())
				if result == gitignore.Exclude {
					LogDebug("Auto-detected %s in %s: filepattern %q matches %q", t, workId, f, path)
					matched = true
					return filepath.SkipAll
				} else if result == gitignore.Include {
					LogDebug("Auto-detected %s in %s: filepattern %q matches %q", t, workId, f, path)
					matched = false
					return filepath.SkipAll
				}
			}

			return nil
		})
	}
	return
}

func (ctx *RunContext) DetectTechnologies(workId string) (detecteds []Technology, err error) {
	results := make(chan Technology, len(ctx.TechnologiesRepository))
	errs := make(chan error, len(ctx.TechnologiesRepository))
	wg := sync.WaitGroup{}

	for _, tech := range ctx.TechnologiesRepository {
		wg.Add(1)
		go func(tech Technology, results chan Technology, errors chan error, wg *sync.WaitGroup) {
			matched, err := tech.Detect(ctx, workId)
			if err != nil {
				errors <- fmt.Errorf("while trying to detect %s: %w", tech, err)
			}
			if matched {
				results <- tech
			}
			wg.Done()
		}(tech, results, errs, &wg)
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			return detecteds, err
		}
	}

	for tech := range results {
		detecteds = append(detecteds, tech)
	}

	sort.Slice(detecteds, func(i, j int) bool {
		return detecteds[i].Slug < detecteds[j].Slug
	})

	return
}

func (t Tag) ReferredToBy(name string) bool {
	return stringsLooselyMatch(name, append(t.Aliases, t.Plural, t.Singular)...)
}

func (ctx *RunContext) FindTag(name string) (result Tag, ok bool) {
	for _, tag := range ctx.TagsRepository {
		if tag.ReferredToBy(name) {
			return tag, true
		}
	}
	return Tag{}, false
}

func (t Technology) ReferredToBy(name string) bool {
	return stringsLooselyMatch(name, append(t.Aliases, t.Slug, t.Name)...)
}

func (ctx *RunContext) FindTechnology(name string) (result Technology, ok bool) {
	for _, tech := range ctx.TechnologiesRepository {
		if tech.ReferredToBy(name) {
			return tech, true
		}
	}
	return Technology{}, false
}

func (ctx *RunContext) DetectTags(workId string, techs []Technology) (detecteds []Tag, err error) {
	results := make(chan Tag, len(ctx.TagsRepository))
	errs := make(chan error, len(ctx.TagsRepository))
	wg := sync.WaitGroup{}

	for _, tag := range ctx.TagsRepository {
		wg.Add(1)
		go func(tag Tag, results chan Tag, errors chan error, wg *sync.WaitGroup) {
			matched, err := tag.Detect(ctx, workId, techs)
			if err != nil {
				errors <- fmt.Errorf("while trying to detect %s: %w", tag, err)
			}
			if matched {
				results <- tag
			}
			wg.Done()
		}(tag, results, errs, &wg)
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			return detecteds, err
		}
	}

	for tag := range results {
		detecteds = append(detecteds, tag)
	}

	sort.Slice(detecteds, func(i, j int) bool {
		return detecteds[i].Plural < detecteds[j].Plural
	})

	return
}

func (ctx *RunContext) LoadTagsRepository() ([]Tag, error) {
	if len(ctx.TagsRepository) > 0 {
		return ctx.TagsRepository, nil
	}

	var tags []Tag
	if ctx.Config.Tags.Repository == "" {
		LogWarning("No tags repository specified in configuration at %s", ctx.Config.source)
		return []Tag{}, nil
	}
	raw, err := readFileBytes(ctx.Config.Tags.Repository)
	if err != nil {
		return []Tag{}, fmt.Errorf("while reading %s: %w", ctx.Config.Tags.Repository, err)
	}

	err = yaml.Unmarshal(raw, &tags)
	if err != nil {
		return []Tag{}, fmt.Errorf("while decoding YAML: %w", err)
	}

	ctx.TagsRepository = tags
	return tags, nil
}

func (ctx *RunContext) LoadTechnologiesRepository() ([]Technology, error) {
	if len(ctx.TechnologiesRepository) > 0 {
		return ctx.TechnologiesRepository, nil
	}

	var technologies []Technology
	if ctx.Config.Technologies.Repository == "" {
		LogWarning("No technologies repository specified in configuration at %s", ctx.Config.source)
		return []Technology{}, nil
	}
	raw, err := readFileBytes(ctx.Config.Technologies.Repository)
	if err != nil {
		return []Technology{}, fmt.Errorf("while reading %s: %w", ctx.Config.Technologies.Repository, err)
	}

	err = yaml.Unmarshal(raw, &technologies)
	if err != nil {
		return []Technology{}, fmt.Errorf("while decoding YAML: %w", err)
	}

	ctx.TechnologiesRepository = technologies
	return technologies, nil
}
