# Extra markdown features

ortfo supports a few additional markdown features:

## Abbreviations

Syntax
: ```md
  *[YAML]: Yet Another Markup Language
  ```

In `database.json`
: Every occurence of `YAML` in paragraphs will be replaced with `<abbr title="Yet Another Markup Language">YAML</abbr>`, directly in the paragraph content[^1]

[^1]: Paragraph content is in `(work).content.(language).blocks.(block).content`. (See [Database format](/db/database-format.md) for more information on where things are stored in the `database.json`)

## Footnotes

Syntax
: `footnote reference[^1]` and then `[^1]: footnote content`

In `database.json`
: Stored in `(work).content.(language).footnotes` as an object, mapping references to content. In this example, the object would be:

    ```json
    {
        "1": "footnote content"
    }
    ```

## Smarty pants

Typographic replacements

| Write | `--` | `---` | `->` | `<-` | `...` | `<<` | `>>` |
| ----- | ---- | ----- | ---- | ---- | ----- | ---- | ---- |
| Get   | –    | —     | →    | ←    | …     | «    | »    |

::: tip
Of course, these replacements are not applied in code blocks or `inline code`.
:::
