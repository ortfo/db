<script setup>
  import schema from '/schemas/latest/tags.schema.json';
</script>

# Tags

Tags are a way to categorize your different projects.

## Usage example

<figure>
  <img src="/examples/tags-index.png"></img>
  <figcaption>Show all works tagged "command-line" <a href="https://ewen.works/command-line">on ewen.works</a></figcaption>
</figure>


## Declaration

You can define all valid tags for your works in a central place with a `tags.yaml` file.

Then, reference the path to that file in your `ortfodb.yaml` configuration file:

```yaml
...
  at: media/

technologies:
  repository: path/to/technologies.yaml

tags: // [!code focus]
  repository: path/to/tags.yaml// [!code focus]
```

### Example

```yaml
# yaml-language-server: $schema=https://ortfo.org/tags.schema.json

- singular: automation
  plural:   automation

- singular: command-line
  plural:   command-line
  aliases:  [cli, command line]
  description: Programs that run in the terminal, a text-based interface for computers.
  learn more at: https://www.hostinger.com/tutorials/what-is-cli
```

::: tip What's that `# yaml-language-server` thing?
See [Declaring JSON Schemas on files](/db/json-schemas.md#on-files-directly)
:::

### Available properties

<JSONSchema :schema :headings="4" type="Tag" />



## Usage

In your work's description file, refer to tag names by their `plural`, `singular` or any of the `aliases`:

```md
---
wip: true
tags: [automation, cli] // [!code focus]
made with: [go, vite]
---

# ortfo
```

## Enforcing correct tags <Badge type=tip text=Planned />

In the future, ortfo/db might enforce that all tags used in your works are defined in the `tags.yaml` file.

This would prevent typos.
