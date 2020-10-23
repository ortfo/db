# portfoliodb

****
> I'm following RDD (readme-driven development) for this project, so, until v0.1.0 is released, this document describes what the program will look like
****

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

## Installation

Pre-compiled binaries are available through [GitHub Releases](https://help.github.com/en/github/administering-a-repository/releasing-projects-on-github):

```shell
$ wget https://github.com/ewen-lbh/portfoliodb/releases/latest/portfoliodb
# Put the command in a directory that is in your PATH, so that you can use portfoliodb from anywhere, e.g.:
$ mv portfoliodb /usr/bin/portfoliodb
```

See [Compiling](#compiling) for instructions on how to compile this yourself

## Usage

See usage [here](./USAGE)

## How it works

Your database is a folder, which has one folder per work in it.
In each folder, you'll have a markdown file describing your work, and other files relevant to the work (images, PDFs, audio files, videos, etc.).

Here's an example tree:

```directory-tree
database/
├── ideaseed
│   ├── logo.png
│   └── description.md
├── phelng
│   └── description.md
├── portfolio
│   └── description.md
└── portfoliodb
    └── description.md
```

"Building" your database is just translating that easy-to-maintain and natural directory tree to a single JSON file, easily consummable by your frontend website. This way, you can add new projects to your portfolio without having to write a single line of code: just create a new folder, describe your project, build the database, upload it, and done!

### `description.md` files

Description files are separated in "blocks": blocks are separated by an empty line.
There are:

- [blocks of text (paragraphs)](#paragraphs)
- [embed declarations (images, videos, audio files, PDFs, etc.)](#media)
- [(isolated) links](#links): useful for linking to other places where the work appears, e.g. a marketplace where to work is sold, a source code repository, etc.

You can also translate your description.md file into multiple languages by using [language markers](#language-markers)

Start your file with a top-level header `# Like this` to give your work a title (it can differ from the folder's name, since the folder name is used as the work's identifier, and is guaranteed to be unique).

#### Paragraphs

These blocks allow you to write some text using an extended markdown syntax, adding support for abbreviations and footnotes.

Paragraphs will be accessible in the JSON file in the `paragraphs` object. Each paragraph has two properties: `content`, which contains the paragraph content, and an `id`, which can be specified manually:

```markdown
other stuff...

(my-paragraph-id)
The start of the paragraph.
Specify the paragraph's ID by starting your paragraph with the ID surrounded by parentheses on a single line.

other stuff...
```

and will be empty otherwise. This `id` can be useful to link to a specific paragraph of your page by using it as, for example, a `<p>` tag's `id`: you can then link to that specific paragraph with `https://example.com/...#my-paragraph-id`

#### Media

A "media" block allows you to declare files embedded in your page: YouTube videos, local files, etc.

Since markdown does not natively support the declaration of embeds other than images, a new syntax was created:

```markdown
>[alt text "title"](./demo.mp4)
```

The syntax is the same as the image's, by replacing the `!` by a `>` (it looks like a play button).

`source` can be a relative path, an absolute one or a URL, just like images.

When building, the compiler will look for these files and analyze them to determine their content type, dimensions, aspect ratio, duration and file size, and will then be accessible in the JSON file as an array of media objects having the following structure:

```json
{
  "dimensions": {
    "height": 1080, // 0 if the file has no dimensions (eg. an audio file)
    "width": 1920, // 0 for the same reasons
    "aspect_ratio": 1.777777778 // 0 if either of the dimensions are zero. aspect_ratio is width / height.
  },
  "source": "./demo.mp4",
  "alt": "alt text",
  "title": "title",
  "duration": 68, // In seconds. 0 if the file has no duration (eg. an image)
  "size": 1854210, // In bytes
  "content_type": "video/mp4", // MIME types
}
```

You _can_ use the `![]()` for images (and anything else in fact), the compiler doesn't care, since it analyzes each file to determine its content type.

#### Links

Of course, you can use links inside of a paragraphs, but you can also declare isolated links that don't need context to be meaningful. Here are some use cases:

- For a website, you can link to the source code repository and the website itself,
- For a t-shirt, you can link to a marketplace so that people viewing that work can buy it
- and plenty of other use cases

## Configuration

Put this in `.portfoliodb.yml` in the root of your database:

```yaml
build steps:
  - step: extract colors
    extract:
      - primary
      - secondary
      - tertiary
    default file name: logo.png

  - step: make gifs
    # <filetitle> refers to the filename without its extension.
    file name template: <filetitle>.gif

  - step: make thumbnails
    widths: [20, 100, 500]
    # Paths are always relative to the work's database folder
    file name template: ../../static/thumbs/<id>/<width>.png


validate:
  checks:
    # can be `off` (not checked for)
    # can be `on` (uses the default level)
    # can be a level:
    # - `fatal`: also checked when building, triggers end of build if fails
    # - `error`: prints an error message (red), makes validate command exit with 1
    # - `warn` : prints a warning message (orange), does not make validate exit with 1
    # - `info` : regular message, informative
    # these are the default values
    schema compliance: fatal
    work folder uniqueness: fatal
    work folder safeness: error
    yaml header: error
    title presence: error
    title uniqueness: error
    tags presence: warn
    tags knowledge: error
    working media: warn
    working urls: off
```

PRO TIP: You can use the provided `.portfoliodb.yml.schema.json` to validate your YAML file
with this JSONSchema

## Extra markdown features

Except for the `>[text](video/audio URL/filepath)` feature, the markdown also supports a number of non-standard features:

- all of what GFM supports (except autolinking of issues and commit hashes, ofc)
- Abbreviations: `*[YAML]: Yet Another Markup Language`
- Definition lists: `- key: value` or the more standard, [PHP-markdown-extra-style](https://michelf.ca/projects/php-markdown/extra/#def-list)
- Admonitions: `!!! type "Optional title"`, see [this documentation](https://python-markdown.github.io/extensions/admonition/)
- Footnotes: `footnote reference[^1]` and then `[^1]: footnote content`
- Markdown in HTML: [See documentation here](https://python-markdown.github.io/extensions/md_in_html/)
- (off by default) New-line-to-line-break: Transforms line breaks in markdown into `<br>`s, see [the documentation](https://python-markdown.github.io/extensions/nl2br/)
- Smarty pants: typographic replacements (not replaced inside code):
  - `--` to –
  - `---` to —
  - `->` to →
  - `<-` to ←
  - `...` to …
  - `<<` to «
  - `>>` to »
- (off by default) Anchored headings: Each headings is assigned an id to reference in the URL with `example.com#heading`

### Configuring markdown

The extra features discussed just above are all available or disable, using the module name:

_.portfoliodb.yml_
```yaml
markdown:
  abbreviations: on
  definition lists: on
  admonitions: off
  footnotes: on
  markdown in html: on
  new-line-to-line-break: on
  smarty pants: off
  anchored headings:
  # you can also use an object form to pass in config options
    enabled: yes
    format: <content> # default value
  custom syntaxes:
    # this is just an example, not an actual implementation of the video/audio embed feature
    - from: '>\[(?P<fallback>[^\]]+)\]\((?P<source>.+)\)'
      to: <video src="${source}">${fallback}</video>
```

# Compiling

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. `make` the binary: `make`
4. Install it (this just copies the file to `/usr/bin/`): `make install`
