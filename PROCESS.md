Processing steps of `description.md` files
=========================================

Initial file
-----------

```markdown
---
yaml: header
---

# name

:: lang-PLACE

(paragraph-id)
paragraph contents

(another-paragraph-id)
the PARA lorem ipsum

![alt "title (optional)"](image source uri)

>[alt "title"](video/audio/pdf/plaintext/other source uri)

[link name](url)

*[PARA]: The abbreviation definition
```

Standard markdown parser
------------------------

Relevant functions:

- `description.go:StandardParse`

```json
{
    "yaml": "header"
}
```

and

```html
<h1>name</h1>

<p>:: lang-PLACE</p>

<p>(paragraph-id)<br/>paragraph contents</p>

<p>(another-paragraph-id)<br/>the PARA lorem ipsum</p>

<p><img alt="alt" title="title (optional)" src="image source uri"/></p>

<p>&gt;[alt "title"](video/audio/pdf/plaintext/other source uri)</p>

<p><a href="url">link name</a></p>

<p>*[PARA]: The abbreviation definition</p>
```

Custom markdown syntaxes
------------------------

Relevant functions:

- `description.go:ParseMediaEmbed`
- `description.go:CollectionAbreviation`
- `description.go:ApplyAbbreviations`
- `description.go:AnchorParagraphs`
- `description.go:ParseLanguageDeclarations`

```json
{
    "yaml": "header"
}
```

and

```html
<h1>name</h1>

<div lang="lang-PLACE">

<p id="paragraph-id">paragraph contents</p>

<p id="another-paragraph-id">the <abbr title="The abbreviation definition</p>">PARA</abbr> lorem ipsum</p>

<p><img alt="alt" title="title (optional)" src="image source uri"/></p>

<p><MEDIA alt="alt" title="title" src="video/audio/pdf/plaintext/other source uri" /></p>

<p><a href="url">link name</a></p>

</div>
```

To JSON object
--------------

Relevant functions:

- `description.go:PseudoHTMLtoWorkObject`

```json
{
    "yaml": "header",
    "name": "name",
    "paragraphs": {
        "lang-PLACE": [
            {
                "id": "paragraph-id",
                "contents": "paragraph contents"
            },
            {
                "id": "another-paragraph-id",
                "contents": "the <abbr title=\"The abbreviation definition</p>\">PARA</abbr> lorem ipsum"
            }
        ]
    },
    "media": {
        "lang-PLACE": [
            {
                "id": "alt",
                "type": "image",
                "source": "image source uri",
                "alt": "alt",
                "title": "title (optional)"
            },
            {
                "id": "alt-1",
                "source": "video/audio/pdf/plaintext/other source uri",
                "alt": "alt",
                "title": "title"
            }
        ]
    },
    "links": {
      "lang-PLACE": [
        {
          "id": "link-name",
          "name": "link name",
          "url": "url"
        }
      ]
    }
}
```

Media filetype detection
-----------------------

Relevant functions:

- `media.go:GetFiletype`

```json
{
    "yaml": "header",
    "name": "name",
    "paragraphs": {
        "lang-PLACE": [
            {
                "id": "paragraph-id",
                "contents": "paragraph contents"
            },
            {
                "id": "another-paragraph-id",
                "contents": "the <abbr title=\"The abbreviation definition</p>\">PARA</abbr> lorem ipsum"
            }
        ]
    },
    "media": {
        "lang-PLACE": [
            {
                "id": "alt",
                "type": "image",
                "format": "png",
                "mime": "image/png",
                "source": "image source uri",
                "alt": "alt",
                "title": "title (optional)"
            },
            {
                "id": "alt-1",
                "source": "video/audio/pdf/plaintext/other source uri",
                "type": "video",
                "format": "mp4",
                "mime": "video/mp4",
                "alt": "alt",
                "title": "title"
            }
        ]
    },
    "links": {
      "lang-PLACE": [
        {
          "id": "link-name",
          "name": "link name",
          "url": "url"
        }
      ]
    }
}
```

Thumbnails generation (build step)
----------------------------------

Relevant functions:

- `media.go:GenerateThumbnails`

```diff
...
                "format": "png",
                "mime": "image/png",
                "source": "image source uri",
                "alt": "alt",
                "title": "title (optional)",
+                 "thumbnails": [
+                     {
+                         "height": 20,
+                         "width": 20,
+                         "aspect_ratio": 1,
+                         "type": "image",
+                         "format": "jpg",
+                         "mime_type": "image/jpg",
+                         "// source": "https://thumbnails.ewen.works/20/20/+ trigonometry-synth/trigonometry-synth.png",
+                         "source": "https://static.ewen.works/works/+ science-is-beautiful/trigonometry/thumbs/20.png"
+                     },
+                     {
+                         "height": 500,
+                         "width": 500,
+                         "aspect_ratio": 1,
+                         "type": "image",
+                         "format": "jpg",
+                         "mime_type": "image/jpg",
+                         "source": "https://static.ewen.works/works/+ science-is-beautiful/trigonometry/thumbs/500.png"
+                     }
+                 ]
+             },
            {
                "id": "alt-1",
                "source": "video/audio/pdf/plaintext/other source uri",
                "type": "video",
                "format": "mp4",
...
```

Color extraction (build step)
-----------------------------

```diff
...
                "format": "png",
                "mime": "image/png",
                "source": "image source uri",
                "alt": "alt",
                "title": "title (optional)",
+                 "colors": {
+                   "primary": "#ebf580",
+                   "secondary": "#c0ffee",
+                 }
                "thumbnails": [
                    {
                        "height": 20,
                        "width": 20,
                        "aspect_ratio": 1,
...
```
