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

```docopt
Usage:
  portfoliodb [options] <database> build <to-filepath> [--config=FILEPATH] [-msS] [--]
  portfoliodb [options] replicate <from-filepath> <to-directory> [--config=FILEPATH]
  portfoliodb [options] <database> add <fullname> [<metadata-item>...]
  portfoliodb [options] <database> validate <database>

Options:
  -C --config=<filepath>      Use the configuration path at <filepath>. [default: .portfoliodb.yml]
  -m --minified               Output a minifed JSON file
  -s --silent                 Do not write to stdout
  -S --scattered              Operate in scattered mode. See Scattered Mode section for more information.

Examples:
  portfoliodb database build database.json
  portfoliodb database add schoolsyst/presentation -#web -#site --color 268CCE
  portfoliodb replicate database.json replicated-database --config=.portfoliodb.yml

Commands:
  build <from-directory> <to-filepath>
    Scan in <from-directory> for folders with description.md files
    (and potential media files)
    and compile the whole database into a JSON file at <to-filepath>

  replicate <from-filepath> <to-directory>
    The reverse operation of 'build'.
    Note that <to-directory> must be an empty directory

  add <name> [<metadata-item>...]
    Creates a new description.md in the appropriate folder.
    <name> is the work's name.
    You can provide additional metadata items in the form --ITEM_NAME=VALUE,
    eg. 'add phelng --tag=cli --tag=program' will generate ./phelng/description.md,
    with the following contents:
    ---
    collection: null
    ---
    # phelng
    program, cli

  validate <database>
    Make sure that everything is OK in the database:
    Each one of these checks are configurable and deactivable in .portfoliodb.yml:validate.checks,
    the step name is the one in [square brackets] at the beginning of these lines.
    1. [schema compliance] validate compliance to schema for .portfoliodb.yml and .portfoliodb-metadata.yml
    2. [work folder names] check work folder names for url-unsafe characters or case-insensitively non-unique folder names
    3. for each work directory:
        a. [yaml header] check YAML header for unknown keys using .portfoliodb-metadata.yml
        b. [title presence] check presence of work title
        c. [title uniqueness] check uniqueness (case-insensitive) of work title
        d. [tags presence] check if at least one tag is present
        e. [tags knowledge] check absence of unknown tags (using .portfoliodb-metadata.yml)
        f. [working media files] check all local paths for links (audio/video files, image files, other files)
        g. [working urls] check that no http url gives errors

Scattered mode:
  With this mode activated, when building, portfoliodb will go through each folder (non-recursively) of <from-directory>, and, if it finds a .portfoliodb file in the folder, consider the files in that .portfoliodb folder.

  Consider the following directory tree:

  <from-directory>
    project1
	  index.html
	  src
	  dist
	  .portfoliodb
	    file1.png
		description.md
	project2
	  .portfoliodb
	    file-2.png
		description.md
	otherfolder
	  stuff

  Running portfoliodb build --scattered on this tree is equivalent to builing without --scattered on the following tree:

  <from-directory>
    project1
	  file.png
	  description.md
	project2
	  file-2.png
	  description.md

  Concretely, it allows you to store your portfoliodb descriptions and supporting files directly in your projects, assuming that your store all of your projects under the same directory.
`
```

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

{#my-paragraph-id}
The start of the paragraph.
Specify the paragraph's ID by starting your paragraph with a {#your-identifier} on a single line.

other stuff...
```

and will be empty otherwise. This `id` can be useful to link to a specific paragraph of your page by using it as, for example, a `<p>` tag's `id`: you can then link to that specific paragraph with `https://example.com/...#my-paragraph-id`

#### Media

A "media" block allows you to declare files embedded in your page: YouTube videos, local files, etc.

With native markdown, you can only declare embeds for _images_. We abuse the syntax to extend it to _any file you want_.

```markdown
![alt text "title"](./demo.mp4)
```

`source` can be a relative path, an absolute one or a URL.

When building, the compiler will look for these files and analyze them to determine their content type, dimensions, aspect ratio, duration and file size, and will then be accessible in the JSON file as an array of media objects having the following structure:

```js
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
    default file names: [logo.png]

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
