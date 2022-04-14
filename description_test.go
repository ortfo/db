package ortfodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDescriptionEmptyParagraphsAreAdded(t *testing.T) {
	ctx := RunContext{}

	actual := ctx.ParseDescription(`---
some: metadata
right: here
---

:: fr

# A title

Some paragraph, an empty one is below, beware!

<p></p>

:: en

# Another title

<p></p>

HAHA!

`)

	expected := ParsedDescription{
		Metadata: map[string]interface{}{
			"some":  "metadata",
			"right": "here",
		},
		Title: map[string]string{
			"fr": "A title",
			"en": "Another title",
		},
		Paragraphs: map[string][]Paragraph{
			"fr": {
				{
					ID:      "",
					Content: "<p>Some paragraph, an empty one is below, beware!</p>",
				},
				{
					ID:      "",
					Content: "<p></p>",
				},
			},
			"en": {
				{
					ID:      "",
					Content: "<p></p>",
				},
				{
					ID:      "",
					Content: "<p>HAHA!</p>",
				},
			},
		},
		MediaEmbedDeclarations: map[string][]MediaEmbedDeclaration{
			"fr": {},
			"en": {},
		},
		Links: map[string][]Link{
			"fr": {},
			"en": {},
		},
		Footnotes: map[string]Footnotes{
			"fr": {},
			"en": {},
		},
	}

	assert.Equal(t, expected, actual)
}
