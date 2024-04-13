# ortfo/db

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

![](./demo.gif)

## Installation

Pre-compiled binaries are available through [GitHub Releases](https://help.github.com/en/github/administering-a-repository/releasing-projects-on-github):

```shell
$ wget https://github.com/ewen-lbh/portfoliodb/releases/latest/ortfodb
# Put the command in a directory that is in your PATH, so that you can use portfoliodb from anywhere, e.g.:
$ mv portfoliodb /usr/bin/ortfodb
```

See [Compiling](#compiling) for instructions on how to compile this yourself

## Usage

```docopt
<<<<USAGE>>>>
```

## How it works

Your database is a folder, which has one folder per work in it.
In each folder, you'll have a markdown file describing your work, and other files relevant to the work (images, PDFs, audio files, videos, etc.).

Here's an example tree:

```directory-tree
database/
├── helloworld
│   ├── logo.png
│   └── description.md
├── resume
│   └── description.md
├── portfolio
│   └── description.md
└── hackernews-clone
    └── description.md
```

"Building" your database is just translating that easy-to-maintain and natural directory tree to a single JSON file, easily consummable by your frontend website. This way, you can add new projects to your portfolio without having to write a single line of code: just create a new folder, describe your project, build the database, upload it, and done!

### "Scattered mode"

If you prefer, you can store your description.md files alongside your projects themselves, instead of having everything in a single folder. This use case is referred to as "Scattered mode". Use the `--scattered` flag to enable it.

This mode expects you to have all of your projects stored in a single directory (with each project being its own folder in that directory). Then, your description.md file (and potentially other resources like screenshots or photos of the work) live in a `.ortfo` folder that's in the projects' folders. The example tree from above becomes:

```directory-tree
projects/
├── helloworld
│   ├── .ortfo
│   │   ├── logo.png
│   │   └── description.md
│   └── main.py
├── resume
│   └── .ortfo
│       └── description.md
├── portfolio
│   └── .ortfo
│       └── description.md
└── hackernews-clone
    └── .ortfo
        └── description.md
```



Of course, your actual project files are still where they are and are left untouched (like the main.py file in the above example)



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

Put this in `ortfodb.yaml` in the root of your database:



## Extra markdown features

- Abbreviations: `*[YAML]: Yet Another Markup Language`

- Footnotes: `footnote reference[^1]` and then `[^1]: footnote content`

- Smarty pants: typographic replacements (not replaced inside code):

  - `--` to –
  - `---` to —
  - `->` to →
  - `<-` to ←
  - `...` to …
  - `<<` to «
  - `>>` to »

## Compiling

The build tool is [Just](https://just.systems), a modern alternative to Makefiles. See [installation](https://github.com/casey/just?tab=readme-ov-file#installation)

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. Compile & install in `~/.local/bin/` `just install`... or simply build a binary to your working directory: `just build`.
