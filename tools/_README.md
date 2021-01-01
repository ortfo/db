# portfoliodb

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

## Installation

Pre-compiled binaries are available through [GitHub Releases](https://help.github.com/en/github/administering-a-repository/releasing-projects-on-github):

```shell
$ wget https://github.com/ewen-lbh/portfoliodb/releases/latest/portfoliodb
# Put the command in a directory that is in your PATH, so that you can use portfoliodb from anywhere, e.g.:
$ mv portfoliodb /usr/bin/portfoliodb
```

See [Compiling](#compiling-it-yourself) for instructions on how to compile this yourself

## How it works

### What's a portfolio anyway?

A portfolio is a collection of 'works' or 'projects' that you showcase.
We'll call them 'works' throughout the documentation.

Thus, a portfolio's database is a simple array of Work objects that contain various information, such as metadata, a title, and the bulk of the article where you describe and talk about your Work, how you created it, etc.

### The data you store

A Work object has the following information stored:

- metadata: an arbitrary object of data, useful for storing auxiliary data like tags, creation date, etc.
- title: The work's title, as displayed to the visitor of your portfolio
- ID: A URL-safe version of the title, useful to construct URLs pointing to your works
- the article itself: it contains different kinds of sections:
  - paragraphs (markdown text)
  - media (images, videos, audio files, PDFs,…)
  - footnotes
  - _isolated_ links: links that are separated from paragraphs, that can point to other sites related to the Work. Use cases are aplenty, here are some:
    - Your Work is a beautiful T-Shirt, and you have a link to the marketplace where it is sold. You want that link to stand out and not get lost in the paragraphs
    - Your Work is a website, and you obviously want people to be able to visit the website itself easily.

### What problem portfoliodb solves

The problem is, you want:

1. To write and manage that information easily
2. To programmatically access those Works, with a database or trough some machine-readable text format

With portfoliodb, you get both at the same time:

1. You write your works' descriptions in markdown files and organize them into folders: one per work
2. You give those folders to portfoliodb, it gives you back a single JSON file, easily parsable by your frontend.

Check out the [example/ directory](./example/) to get a taste of what it looks like.

### Two ways to place your description files

You can either place your description files directly in your projects' folders, or put them all in one folder

#### Scattered mode

If you are already storing all of your projects inside a single directory, say a `projects` folder, this mode will make using portfoliodb even more enjoyable. Here's how to put your projects into your portfolio:

1. Go to a project's folder, inside your `projects` directory.
2. Create a `.portfoliodb` folder.
3. Go to that new folder
4. Write your `description.md` file

Note that the project's folder name will be used as the Work's ID.

#### Centralized mode

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

### The description file

Now let's get real. Here's how you write those description files.

#### It's just markdown

As the file name suggests, description files are written in [Markdown](https://daringfireball.net/projects/markdown/syntax), the plain-text rich text format we all know and love.

portfoliodb adds in a few popular features to the basic markdown syntax:

- tables
- 'fenced' code blocks (with ```)
- auto-linking of URLs
- striketrough text (`~~like this~~`)
- IDs to headings (useful for linking to a specific section):
  - automatically (based on the heading text)
  - manually, by adding `{#your-custom-id}` above your heading line
- definition lists (using [PHP markdown extra's syntax](https://catalog.olemiss.edu/help/markdown/extra#def-list))
- LaTeX math (`$like_\text{this}$`)
- footnotes (using [Pandoc's syntax](https://garrettgman.github.io/rmarkdown/authoring_pandoc_markdown.html#footnotes))
- hard line breaks: a line break results in a `<br>`, breaking the line as you intended. If you don't like this, it can be disabled.

Except where noted, additional features' syntax are [github's](https://guides.github.com/features/mastering-markdown/)

#### But... with some more liberties

- **Media embeds** If you have some files to show that are not images, we've got you covered. The image embed syntax (`![alt text "title"](source)`) has been extended so that `source` can be any file you want. But if you use an editor that shows you a preview as you type, it'll try to show your non-image file as an image. Not nice. That's why we've introduced an alternative syntax: `>[alt text "title"](source)`. It'll show as a quoted link from any markdown editor, but stil get interpreted as a media embed by portfoliodb. You can use both syntaxes, portfoliodb does not make any difference between the two.
- **Language markers** If you need to translate your portfolio into multiple languages (like your home country's and english), we've got you covered too. You can split your description into multiple languages. More details in [Internationalisation](#internationalisation).

## Let's get technical

### Usage

```
<<<USAGE>>>
```

### Compiling it yourself

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. `make` the binary: `make`
4. Install it (this just copies the file to `/usr/bin/`): `make install`
