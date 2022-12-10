package ortfodb

func FindMedia(works []AnalyzedWork, mediaEmbed MediaEmbedDeclaration) (found bool, media Media) {
	for _, w := range works {
		for _, wsl := range w.Localized {
			for _, b := range wsl.Blocks {
				if b.Type == "media" && b.RelativeSource == mediaEmbed.Source {
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

func (work AnalyzedWork) Languages() []string {
	langs := make([]string, 0, len(work.Localized))
	for lang := range work.Localized {
		langs = append(langs, lang)
	}
	return langs
}
