package ortfodb

func FindMedia(works []Work, mediaEmbed MediaEmbedDeclaration) (found bool, media Media) {
	for _, w := range works {
		for _, ms := range w.Media {
			for _, m := range ms {
				if m.Source == mediaEmbed.Source {
					return true, m
				}
			}
		}
	}
	return
}

func FindWork(works []Work, id string) (found bool, work Work) {
	for _, w := range works {
		if w.ID == id {
			return true, w
		}
	}
	return
}
