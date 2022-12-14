package ortfodb

func FindMedia(works []AnalyzedWork, mediaEmbed Media) (found bool, media Media) {
	for _, w := range works {
		for _, wsl := range w.Content {
			for _, b := range wsl.Blocks {
				if b.Type == "media" && b.RelativeSource == mediaEmbed.RelativeSource {
					return true, b.Media
				}
			}
		}
	}
	return
}

func FindWork(works []AnalyzedWork, id string) (found bool, work AnalyzedWork) {
	for _, w := range works {
		if w.ID == id {
			return true, w
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
