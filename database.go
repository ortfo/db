package ortfodb

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"time"
)

func LoadDatabase(at string, skipValidation bool) (database Database, err error) {
	json := jsoniter.ConfigFastest
	content, err := readFileBytes(at)
	if err != nil {
		return
	}
	if !skipValidation {
		validated, validationErrors, err := validateWithJSONSchema(string(content), DatabaseJSONSchema())
		if err != nil {
			return database, err
		}
		if !validated {
			DisplayValidationErrors(validationErrors, "database JSON")
			err = errors.New("database JSON is invalid")
			return database, err
		}
	}
	err = json.Unmarshal(content, &database)
	return
}

func FindMedia(works Database, mediaEmbed Media, workID string) (found bool, media Media, builtAt time.Time) {
	for _, w := range works {
		if w.ID != workID {
			continue
		}
		for _, wsl := range w.Content {
			for _, b := range wsl.Blocks {
				if b.Type == "media" && b.RelativeSource == mediaEmbed.RelativeSource {
					builtAt, err := time.Parse(time.RFC3339, w.BuiltAt)
					if err != nil {
						return false, media, builtAt
					}
					return true, b.Media, builtAt
				}
			}
		}
	}
	return
}

// FirstParagraph returns the first paragraph content block of the given work in the given language
func (work AnalyzedWork) FirstParagraph(lang string) (found bool, paragraph ContentBlock) {
	for _, block := range work.Content[lang].Blocks {
		if block.Type == "paragraph" {
			return true, block
		}
	}
	return
}
