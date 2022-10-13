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

HAHA![^1]

[^1]: By the way, this shouldn't have trailing <br>'s!

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
					Content: "<p>HAHA!<sup class=\"footnote-ref\" id=\"fnref:1\"><a href=\"#fn:1\">1</a></sup></p>",
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
			"en": {
				"1": "By the way, this shouldn’t have trailing <br/>’s!",
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestImageCaptions(t *testing.T) {
	_, _, actual, _, _, _ := ParseSingleLanguageDescription(`![some alt text "right “there" “here”](https://example.com/source "the title “here”")`)
	expected := []MediaEmbedDeclaration{
		{
			Alt:    `some alt text “right “there” “here”`,
			Title:  `the title “here”`,
			Source: "https://example.com/source",
			Attributes: MediaAttributes{
				Controls: true,
			},
		},
	}
	assert.Equal(t, expected, actual)

}
