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
paragraph contents with _italic text_ and **bold text** that also has `code` (inline) and [a link](https://and-its-url.com/)

(another-paragraph-id)
the PARA lorem ipsum, and footnote[1]

![alt "title (optional)"](image source uri)

>[alt "title"](video/audio/pdf/plaintext/other source uri)

[link name](url)

*[PARA]: The abbreviation definition
[1]: A footnote reference
```

Abbreviations & footnotes collection
------------------------------------

Relevant functions:

- `description.go:CollectAbbreviation`
- `description.go:CollectFootnote`

```diff
  ---
  yaml: header
  ---

  # name

  :: lang-PLACE

  (paragraph-id)
  paragraph contents with _italic text_ and **bold text** that also has   `code` (inline) and [a link](https://and-its-url.com/)

  (another-paragraph-id)
  the PARA lorem ipsum, and footnote[1]

  ![alt "title (optional)"](image source uri)

  >[alt "title"](video/audio/pdf/plaintext/other source uri)

  [link name](url)
-
- *[PARA]: The abbreviation definition
- [1]: A footnote reference
```

Collected abbreviations

```go
type Abbreviation struct {
  Name string
  Definition string
}

collected := [
  Abbreviation{Name: "PARA", Definition: "The abbreviation definition"}
]
```

Collected footnotes

```go
type Footnote struct {
  Number uint16 // Oh no, what a bummer, you can't have more than 65 535 footnotes
  Content string
}

collected := [
  Footnote{Number: 1, Content: "A footnote reference"}
]
```

To JSON object
--------------

Relevant functions:

- `description.go:ExtractName`
- `description.go:ExtractParagraphs`
- `description.go:ExtractMedia`
- `description.go:ExtractLink`
- `description.go:SplitOnLanguageMarkers`

```json
{
    "yaml": "header",
    "name": "name",
    "paragraphs": {
        "lang-PLACE": [
            {
                "id": "paragraph-id",
                "contents": "paragraph contents with _italic text_ and **bold text** that also has `code` (inline) and [a link](https://and-its-url.com/)"
            },
            {
                "id": "another-paragraph-id",
                "contents": "the PARA lorem ipsum, and footnote[1]"
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

Footnotes & abbreviations processing
------------------------------------

Relevant functions:

- `description.go:ApplyAbbreviation`
- `description.go:ApplyFootnote`

```diff
      "paragraphs": {
          "lang-PLACE": [
              {
                  "id": "paragraph-id",
                  "contents": "paragraph contents with _italic text_ and **bold text** that also has `code` (inline) and [a link](https://and-its-url.com/)"
              },
              {
                  "id": "another-paragraph-id",
-                 "contents": "the PARA lorem ipsum, and footnote[1]"
+                 "contents": "the <abbr title=\"The abbreviation definition\">PARA</abbr> lorem ipsum, and footnote<a href=\"#footnote-1\" title=\"A footnote reference\" id=\"footnote-1-ref-1\"><sup>1</sup></a>"
              }
          ]
```

Markdown to HTML for paragraphs' content
-----------------------

Relevant functions:

- `media.go:GetFiletype`

```diff
...
     "paragraphs": {
         "lang-PLACE": [
             {
                 "id": "paragraph-id",
-                "contents": "paragraph contents with _italic text_ and **bold text** that also has `code` (inline) and [a link](https://and-its-url.com/)"
+                "contents": "paragraph contents with <em>italic text</em> and <strong>bold text</strong> that also has <code>code</code> (inline) and <a href="https://and-its-url.com/">a link &quot;with a title&quot;</a>"
             },
...
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

Relevant functions:

- `media.go:ExtractColors`

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
+                 },
                  "thumbnails": [
                      {
                          "height": 20,
                          "width": 20,
                          "aspect_ratio": 1,
...
```
