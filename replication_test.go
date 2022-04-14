package ortfodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyParagraphsAreReplicated(t *testing.T) {
	ctx := RunContext{}
	expected := ctx.ParseDescription(`---
right: here
some: metadata
---


:: fr

# A title

Some paragraph, an empty one is below, beware!

<p></p>



:: en

# Another title

<p></p>

HAHA!`)

	actual, err := ReplicateDescription(ParsedDescription{
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
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, ctx.ParseDescription(actual))
}
