# Your First `description.md` file


## Where do I put it???

Your database is a folder, which has one folder per work in it.
In each folder, you'll have a markdown file, called `description.md`, describing your work, and other files relevant to the work (images, PDFs, audio files, videos, etc.).

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

::: tip
If you prefer to put your description files alongside your projects themselves instead of having everything in a central folder somewhere else, check out [Scattered mode](/db/scattered-mode).
:::

## Creating a `description.md` file quickly with some pre-filled metadata

You can use the [`ortfodb add`](/db/commands/add) command to quickly create a new description.md file. Some metadata will be pre-determined: for example, the creation date will default to the last git commit's date, if the project you're referencing is a git repository.

<video src="/db/demo-add.mp4" controls autoplay muted />

## What do I write in it?

### An example

Here's an example of a typical `description.md` file:

```md
---
started: 2023-04-12
tags: [web, design, ux, ui]
made with: [figma, react, go]
---

# My awesome project

This is a paragraph of text. It can contain **bold**, *italic*, and [links](https://example.com).

![](./demo.mp4 "Some caption")

[Link to the source code](https://github.com/ortfo/db)
```

### Blocks

Description files are separated in "blocks": blocks are separated by an empty line.
There are:

- [blocks of text (paragraphs)](#paragraphs)
- [embed declarations (images, videos, audio files, PDFs, etc.)](#media)
- [(isolated) links](#links): useful for linking to other places where the work appears, e.g. a marketplace where to work is sold, a source code repository, etc.

You can also translate your description.md file into multiple languages by using [language markers](/db/internationalization#language-markers).

Start your file with a top-level header `# Like this` to give your work a title (it can differ from the folder's name, since the folder name is used as the work's identifier, and is guaranteed to be unique).

#### Paragraphs


```markdown{9}
---
started: 2023-04-12
tags: [web, design, ux, ui]
made with: [figma, react, go]
---

# My awesome project

This is a paragraph of text. It can contain **bold**, *italic*, and [links](https://example.com).

![](./demo.mp4 "Some caption")

[Link to the source code](https://github.com/ortfo/db)
```

These blocks allow you to write some text using an extended markdown syntax, adding support for abbreviations and footnotes.

#### Media

```markdown{11}
---
started: 2023-04-12
tags: [web, design, ux, ui]
made with: [figma, react, go]
---

# My awesome project

This is a paragraph of text. It can contain **bold**, *italic*, and [links](https://example.com).

![](./demo.mp4 "Some caption")

[Link to the source code](https://github.com/ortfo/db)
```

A "media" block allows you to declare files embedded in your page: YouTube videos, local files, etc.

With native markdown, you can only declare embeds for _images_. We abuse the syntax to extend it to _any file you want_.

```markdown
![alt text "title"](./demo.mp4)
```

`source` can be a relative path, an absolute one or a URL.

When building, the compiler will look for these files and analyze them to determine useful metadata such as the dimensions, the duration, whether the media has sound, etc.

#### Links

```markdown{13}
---
started: 2023-04-12
tags: [web, design, ux, ui]
made with: [figma, react, go]
---

# My awesome project

This is a paragraph of text. It can contain **bold**, *italic*, and [links](https://example.com).

![](./demo.mp4 "Some caption")

[Link to the source code](https://github.com/ortfo/db)
```

Of course, you can use links inside of a paragraphs, but you can also declare isolated links that don't need context to be meaningful. Here are some use cases:

- For a website, you can link to the source code repository and the website itself,
- For a t-shirt, you can link to a marketplace so that people viewing that work can buy it
- and plenty of other use cases

Having links by themselves allows you to put emphasis on them (maybe display them as buttons), or even display them in a different place than the rest of the content.
