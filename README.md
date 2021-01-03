# portfoliodb

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

## Contents

- [Contents](#contents)
- [Installation](#installation)
- [An introduction](#an-introduction)
  - [What's a portfolio anyway?](#whats-a-portfolio-anyway)
- [What problem portfoliodb solves](#what-problem-portfoliodb-solves)
  - [The data you store](#the-data-you-store)
- [The output file](#the-output-file)
- [Two ways to place your description files](#two-ways-to-place-your-description-files)
  - [Scattered mode](#scattered-mode)
  - [Centralized mode](#centralized-mode)
- [The description file](#the-description-file)
  - [It's just markdown](#its-just-markdown)
  - [But... with some more liberties](#but-with-some-more-liberties)
  - [Okay, but what does it look like?](#okay-but-what-does-it-look-like)
- [Internationalisation](#internationalisation)
- [Media files](#media-files)
  - [The `media` object](#the-media-object)
    - [Textual properties](#textual-properties)
    - [Size, dimensions and duration](#size-dimensions-and-duration)
    - [Attributes, `online` and `has_sound`](#attributes-online-and-has_sound)
  - [Media attributes](#media-attributes)
- [Let's get technical](#lets-get-technical)
  - [Usage](#usage)
  - [Compiling it yourself](#compiling-it-yourself)

## Installation

Pre-compiled binaries are available through [GitHub Releases](https://help.github.com/en/github/administering-a-repository/releasing-projects-on-github):

```shell
$ wget https://github.com/ewen-lbh/portfoliodb/releases/latest/portfoliodb
# Put the command in a directory that is in your PATH, so that you can use portfoliodb from anywhere, e.g.:
$ mv portfoliodb /usr/bin/portfoliodb
```

See [Compiling](#compiling-it-yourself) for instructions on how to compile this yourself

## An introduction

### What's a portfolio anyway?

A portfolio is a collection of 'works' or 'projects' that you showcase.
We'll call them 'works' throughout the documentation.

Thus, a portfolio's database is a simple array of Work objects that contain various information, such as metadata, a title, and the bulk of the article where you describe and talk about your Work, how you created it, etc.

## What problem portfoliodb solves

The problem is, you want:

1. To write and manage that information easily
2. To programmatically access those Works, with a database or trough some machine-readable text format

With portfoliodb, you get both at the same time:

1. You write your works' descriptions in markdown files and organize them into folders: one per work
2. You give those folders to portfoliodb, it gives you back a single JSON file, easily parsable by your frontend.

Check out the [example/ directory](./example/) to get a taste of what it looks like.

### The data you store

A Work object has the following information stored:

- `metadata`: an arbitrary object of data, useful for storing auxiliary data like tags, creation date, etc.
- `title`: The work's title, as displayed to the visitor of your portfolio
- `id`: A URL-safe version of the title, useful to construct URLs pointing to your works
- the article itself: it contains different kinds of sections:
  - `paragraphs` (markdown text)
  - `media` (images, videos, audio files, PDFs,…)
  - `footnotes`
  - _isolated_ `links`: links that are separated from paragraphs, that can point to other sites related to the Work. Use cases are aplenty, here are some:
    - Your Work is a beautiful T-Shirt, and you have a link to the marketplace where it is sold. You want that link to stand out and not get lost in the paragraphs
    - Your Work is a website, and you obviously want people to be able to visit the website itself easily.

## The output file

portfoliodb will give you back a single JSON file containing all of your works.
Here's what it looks like: _(Don't worry, the How is explained later. I think that having the output format in your head before reading the documentation helps)_

```json
[
  {
    "id": "your-works-id-is-its-folder-name",
    "metadata": {
      "any": "metadata",
      "you": "store",
      "in": "the YAML header of the description file"
    },
    "title": {
      "en": "The english title",
      "fr": "The french title"
    },
    "paragraphs": {
      "en": [
        {
          "id": "optional-anchor-for-this-paragraph",
          "content": "This paragraph, converted to HTML"
        }
      ],
      "fr": [
        {
          "id": "",
          "content": "The first paragraph after a :: fr marker"
        }
      ]
    },
    "media": {
      "en": [
        {
          "id": "generated-from-the-alt-text-if-non-empty-else-basename-of-the-source",
          "alt": "alt text",
          "title": "a title",
          "source": "/absolute/path/to/the/file/or/a/URL",
          "content_type": "type/subtype (a MIME type)",
          "size": 7048,
          "dimensions": {
              "width": 500,
              "height": 281,
              "aspect_ratio": 1.779359
          },
          "duration": 344,
          "online": false,
          "attributes": {
            "loop": false,
            "controls": true,
            "playsinline": false,
            "autoplay": false,
            "muted": false
          },
          "has_sound": true
        }
      ],
      "fr": [
        ...
      ]
    },
    "links": {
      "en": [
        {
          "id": "generated-from-the-alt-text",
          "alt": "the alt text",
          "title": "the title",
          "url": "https://example.com"
        }
      ]
    },
    "footnotes": {
      "en": [
        {
          "name": "the-name-by-which-the-footnote-is-referenced",
          "content": "The footnote's text"
        }
      ]
    }
  }
]
```

Learn about:

- [Why there's a `"en": {` and `"fr": {` everywhere or what the hell `:: fr marker` means](#internationalisation)
- [How the `id` is computed](#two-ways-to-place-your-description-files)
- [The `media` object](#media-files)
- [How portfoliodb scans the markdown description file into `links`, `title`, `media`, `paragraphs`, etc.](#the-description-file)

Note: A [JSON schema](https://raw.githubusercontent.com/ortfo/portfoliodb/master/database.schema.json) is included.

## Two ways to place your description files

You can either place your description files directly in your projects' folders, or put them all in one folder

### Scattered mode

If you are already storing all of your projects inside a single directory, say a `projects` folder, this mode will make using portfoliodb even more enjoyable. Here's how to put your projects into your portfolio:

1. Go to a project's folder, inside your `projects` directory.
2. Create a `.portfoliodb` folder.
3. Go to that new folder
4. Write your `description.md` file

Note that the project's folder name will be used as the Work's ID.

### Centralized mode

If your organization is more complex, you don't have all of your projects inside a single directory (or simply don't want to add `.portfoliodb` folders in every project), you can do this:

1. Create a new folder that will contain all of your portfolio's Works (`portfolio-database`, for example)
2. For every Work you wish to add to your portfolio, create a folder whose name is the Work's ID.
3. Write your `description.md` file in that folder.

In both cases, you'll notice that the uniqueness of IDs is guaranteed because the ID is just the folder's name, and, inside a directory, two different folders cannot have the same name.

Once you have written your description files, just tell portfoliodb where to look for `description.md` files, indicating which mode you choose, so that it knows if the files are within `.portfoliodb` folders or not:

- In scattered mode
  ```
  portfoliodb <your big projects directory where you store all of your projects> --scattered build <the output JSON file's name>
  ```
  Example:

  ```
  portfoliodb ~/projects --scattered build portfolio-database.json
  ```

- In centralized mode
  ```
  portfoliodb <your portfolio database directory where you wrote all the description files> build <the output JSON file's name>
  ```
  Example:

  ```
  portfoliodb ~/portfolio/database build portfolio-database.json
  ```

## The description file

Now let's get real. Here's how you write those description files.

### It's just markdown

As the file name suggests, description files are written in [Markdown](https://daringfireball.net/projects/markdown/syntax), the plain-text rich text format we all know and love.

portfoliodb adds in a few popular features to the basic markdown syntax:

- A YAML header (surrounded by two `---` lines, atop the file)
  — This is used to specify [your custom metadata](#the-data-you-store)
- tables
- 'fenced' code blocks (with ```)
- auto-linking of URLs
- striketrough text (`~~like this~~`)
- IDs to headings (useful for linking to a specific section):
  - automatically (based on the heading text)
  - manually, by adding `{#your-custom-id}` above your heading line
- definition lists (using [PHP markdown extra's syntax](https://catalog.olemiss.edu/help/markdown/extra#def-list))
- LaTeX math (`$like_\text{this}$`)
- [footnotes](#the-data-you-store) (using [Pandoc's syntax](https://garrettgman.github.io/rmarkdown/authoring_pandoc_markdown.html#footnotes))
- hard line breaks: a line break results in a `<br>`, breaking the line as you intended. If you don't like this, it can be disabled in [the configuration file](#configuration)

Except where noted, additional features' syntax are [github's](https://guides.github.com/features/mastering-markdown/)

### But... with some more liberties

- **Media embeds** If you have some files to show that are not images, we've got you covered. The image embed syntax (`![alt text "title"](source)`) has been extended so that `source` can be any file you want. But if you use an editor that shows you a preview as you type, it'll try to show your non-image file as an image. Not nice. That's why we've introduced an alternative syntax: `>[alt text "title"](source)`. It'll show as a quoted link from any markdown editor, but stil get interpreted as a media embed by portfoliodb. You can use both syntaxes, portfoliodb does not make any difference between the two.
- **Media attributes** for media embeds: If you terminate your alt text with some special characters, you add some attributes to the media: should the video loop? should it play automatically? You'll likely use those to add attributes to the corresponding `<audio>` or `<video>` tag on your website. More details in [Media attributes](#media-attributes)
- **Language markers** If you need to translate your portfolio into multiple languages (like your home country's and english), we've got you covered too. You can split your description into multiple languages. More details in [Internationalisation](#internationalisation).

### Okay, but what does it look like?

Here's one example that illustrates all of portfoliob's syntax features:

```markdown
---
some: metadata
here is:
  - more
  - metadata
this is: good ol' YAML syntax
---

# ortfo

:: fr

{#premier-paraphe}
## Le premier paragraphe
... Est vraiment super intéréssant.

![some isolated media "with its title" >~](../a-video-file.mp4)

[A link, all alone](https://example.com/)

![a youtube embed, as a media file](https://youtu.be/k43WtSBPeko)

:: en

The english translation of the above.

>[a media, this time with the "quoted link" syntax](../my-superb-recording.flac)
```

As you can see, it's pretty much all markdown, only the [language marker syntax](#internationalisation) seems foreign.

## Internationalisation

If you need to translate your portfolio into multiple languages, portfoliodb provides you a syntax to split your description file into multiple descriptions of the same work, in different languages. Here's what it looks like:

```markdown
---
some: metadata
---

# My title

:: fr

A part marked with language 'fr'

:: en

A part marked with language 'en'
```

Anything **before the first** _language marker_ is considered to be the same in all languages. This allows you to specify your title (and maybe more) only once, if you don't need to translate it.

If the description file contains no languages but other description files do, portfoliodb will consider that this file has the exact same content for all the languages.

However, if you really don't have **any** language markers appearing throughout your whole database, the language of the entire portfolio will be called 'default'. In other words, not putting language markers anywhere is the same as starting each description file with `:: default` (after the metadata section).

In [the output JSON file](#the-output-file), each work's properties that need translation is split into a version for each language: `paragraphs` is not an array of Paragraph objects but an object mapping a language to its array of Paragraph objects:

```json
{
  ...
  "paragraphs": {
    "language": [
      {
        "id": ...,
        "content": ...
      }
    ]
  },
  ...
}
```

this goes for the following properties:

- paragraphs
- links
- media
- title
- footnotes

## Media files

### The `media` object

Just like paragraphs, links or titles, media embeds deserve [internationlization](#internationalisation) too: even if you serve exactly the same files on every language, alt texts and titles still need translation. Therefore, the `media` object maps a language code to an array of Mediæ. Let's see what properties portfoliodb gives you for each media embed:

#### Textual properties

_Strings_

The three pieces of data you declare in your description file are available to you:

```markdown
>[alt "title"](source)
```

are stored respectiveely in `alt`, `title` and `source`. Of course, `alt` and `title` can be omitted, which results in their properties being empty (`""`), _not `null`_.

An `id` is also derived, by [slugifying](https://en.wikipedia.org/wiki/Clean_URL#Slug) the alt text or, if this one is empty, by taking the [basename](https://en.wikipedia.org/wiki/Basename) of the `souce` and removing the extension.

#### Size, dimensions and duration

_Integers_

When the media's source points to an online resource (i.e. when [`online`](#attributes-online-and-has_sound) is `true`), all properties are set to 0. Portofliodb could work them out, but [this is not the case yet](https://github.com/ortfo/portfoliodb/issues/20)

When a property doesn't make sense for a file, its value is set to 0.

| Property                  | Unit    | Makes sense for file types | Description              |
| ------------------------- | ------- | -------------------------- | ------------------------ |
| `dimensions.height`       | pixels  | `image`, `video`           |
| `dimensions.width`        | pixels  | `image`, `video`           |
| `dimensions.aspect_ratio` | Ø       | `image`, `video`           | `width / height`         |
| `duration`                | seconds | `video`, `audio`           |
| `size`                    | bytes   | (all)                      | Size of the file on disk |

#### Attributes, `online` and `has_sound`

_Booleans_

-  `online` indicates whether `souce` points to an online resource (i.e. if it's a URL) or to a local file.
- `has_sound` is set to `true` if and only if the media is either:
  - an audio file
  - a video file that has at least one audio stream.<br>
    ⚠ _This means that if the video file has an audio stream that is complete silence, `has_sound` will be still `true`! See https://github.com/ortfo/portfoliodb/issues/21_
- `attributes` is an object of booleans that provide various indications on how the media should be played. See [Media attributes](#media-attributes) for more information

### Media attributes

Let's imagine that you made a spinner for somebody that you want in your portfolio. The animation lasts for 5 seconds, has no sound and you probably want it to loop. Here, it's reasonable to want your video to autoplay and loop on your website. Here's [a concrete example of what I'm talking about](https://en.ewen.works/legmask-spinner).

You can specify this easily in your markdown without having to restort to HTML, categorizing your media as a paragraph, to portfoliodb's eyes. The idea is that, when special characters are added at the end of a media's alt text, it's interpreted as turning on or off some attributes. Here's a helpful table:

| Character | What it does                                     | Why I chose this one                                                      | Notes                                                                                                                                                                                                                                                                                                                                                                  |
| --------- | ------------------------------------------------ | ------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `~`       | Sets `loop` to `true`                            | It's the closest ASCII (i.e. typable on any keyboard) approximation of ∞. |
| `>`       | Sets both `autoplay` and `muted` to `true`       | Looks like a play button                                                  | [Most browsers will block `autoplay` if the file has sound and no user interaction happened beforehand](https://developer.mozilla.org/en-US/docs/Web/Media/Autoplay_guide#Autoplay_availability), so you cannot set one without setting the other. Either way, autoplaying sound on the web is a dick move. Don't do it. People will get off your beautiful portfolio. |
| `=`       | Does `playsinline = true` and `controls = false` | Makes me think of 'fullscreen'                                            | [playsinline](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/video#attr-playsinline) will play the video inline on mobile. [controls](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/video#attr-controls) is `true` by default, this sets it to `false`.                                                                                         |

Note that portfoliodb spits out JSON files, so this will not result in any HTML whatsoever. There's a one-to-one mapping with HTML attributes because I believe most will use these to control HTML attributes on their portfolio website. But you could totally do something different.

## Let's get technical

### Usage

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
  Concretely, it allows you to store your portfoliodb descriptions and supporting files directly in your projects, assuming that your store all of your projects under the same directory. See the documentation for a more complete explanation.
`
```

### Compiling it yourself

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. `make` the binary: `make`
4. Install it (this just copies the file to `/usr/bin/`): `make install`
