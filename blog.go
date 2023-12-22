package ortfodb

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/bmatcuk/doublestar/v4"
)

type BlogMetadata struct {
	Aliases            []string                      `json:"aliases"`
	Projects           []string                      `json:"projects"`
	Date               time.Time                     `json:"date"`
	Tags               []string                      `json:"tags"`
	Thumbnail          FilePathInsidePortfolioFolder `json:"thumbnail"`
	Private            bool                          `json:"private"`
	AdditionalMetadata map[string]interface{}        `mapstructure:",remain" json:"additional_metadata"`
}

type AnalyzedBlogPage struct {
	ID       string                      `json:"id"`
	BuiltAt  time.Time                   `json:"built_at"`
	Hash     string                      `json:"hash"`
	Metadata BlogMetadata                `json:"metadata"`
	Content  map[string]LocalizedContent `json:"content"`
	Partial  bool                        `json:"partial"`
}

type Blog map[string]AnalyzedBlogPage

// GatherPages gathers all the .md files in inside/**/*.md and returns these paths
func GatherPages(inside string) (paths []string, err error) {
	return doublestar.Glob(os.DirFS(inside), "**/*.md")
}

func BuildBlog(inside string) (blog Blog, err error) {
	ctx := RunContext{}

	paths, err := GatherPages(inside)
	if err != nil {
		err = fmt.Errorf("while gathering pages for blog in %s: %w", inside, err)
	}

	blog = make(Blog)
	for _, path := range paths {
		page, err := ctx.BuildBlogPage(inside, path)
		if err != nil {
			return blog, fmt.Errorf("while building blog page %s: %w", path, err)
		}
		blog[page.ID] = page
	}
	return
}

// BuildBlogPage builds the given blog page at the given path
func (ctx *RunContext) BuildBlogPage(inside string, path string) (page AnalyzedBlogPage, err error) {
	raw, err := os.ReadFile(filepath.Join(inside, path))
	if err != nil {
		err = fmt.Errorf("while reading %s: %w", path, err)
		return
	}
	page.ID = filepath.Join(filepath.Dir(path), strings.TrimSuffix(filepath.Base(path), ".md"))

	hashBytes := md5.Sum(raw)
	hash := base64.StdEncoding.EncodeToString(hashBytes[:])
	page.Hash = hash

	metadata, blocks, title, footnotes, _ := ParseDescription[BlogMetadata](ctx, string(raw))
	page.BuiltAt = time.Now()
	page.Metadata = metadata
	page.Content = make(map[string]LocalizedContent)
	for language, blocksOfLang := range blocks {
		page.Content[language] = LocalizedContent{
			Blocks:    blocksOfLang,
			Title:     title[language],
			Footnotes: footnotes[language],
		}
	}
	return
}
