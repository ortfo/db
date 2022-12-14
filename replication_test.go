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

	actual, err := ReplicateDescription(AnalyzedWork{
		Metadata: WorkMetadata{AdditionalMetadata: map[string]interface{}{
			"some":  "metadata",
			"right": "here",
		},
			Aliases: []string{"alias1", "alias2"},
		},
		Content: map[string]LocalizedWorkContent{
			"fr": {
				Title: "A title",
				Blocks: []ContentBlock{
					{
						Paragraph: Paragraph{
							Content: "Some paragraph, an empty one is below, beware!",
						},
					},
					{
						Paragraph: Paragraph{
							Content: "",
						},
					},
				},
			},
			"en": {
				Title: "Another title",
				Blocks: []ContentBlock{
					{
						Paragraph: Paragraph{
							Content: "",
						},
					},
					{
						Paragraph: Paragraph{
							Content: "HAHA!",
						},
					},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, ctx.ParseDescription(actual))
}
