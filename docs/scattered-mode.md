---
prev: false
---

# Scattered mode

If you prefer, you can store your description.md files alongside your projects themselves, instead of having everything in a single folder. This use case is referred to as "Scattered mode". Use the `--scattered` global flag to enable it.

This mode expects you to have all of your projects stored in a single directory (with each project being its own folder in that directory). Then, your description.md file (and potentially other resources like screenshots or photos of the work) live in a `.ortfo` folder that's in the projects' folders.


Take the following example tree:

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

Using scattered mode, that tree would instead look like this:

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

## Advantages

It's useful if you want to re-use files that you use in the project itself without having to copy anything over.

For example, if you're describing a music album, you most media files that you'll include in the description.md file are probably already in the project's folder. In scattered mode, you can just reference them easily with `../`. For example, if you want to reference `ame-to-yuki-final-v7-final-final.flac`[^1] in your description.md file, you can just write `../ame-to-yuki-final-v7-final-final.flac`.

[^1]: least unhiged music producer file naming scheme

## Configuration

You can change the folder name to use something else, instead of `.ortfo`. Just change the `scattered mode folder` setting in your `ortfodb.yaml` configuration file.

```yaml
...
build metadata file: .lastbuild.yaml
media:
    at: media/
scattered mode folder: .ortfo // [!code focus]
tags:
  repository: ...
```
