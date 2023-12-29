package ortfodb

import "time"

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
