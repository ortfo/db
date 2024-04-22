<script setup>
  import schema from '/schemas/latest/technologies.schema.json';
</script>

# Technologies

ortfo/db has a concept of _technologies_. This allows you to specify with what tools the work was made. Note that the term is intended to be very broad-reaching: if you are describing a painting, for example, “oil paint” and “canvas” would be considered technologies. If you are describing a website, “HTML” and “CSS” would be considered technologies.

## Usage example

<div class="side-by-side">
  <figure>
    <img src="/examples/technologies.png"></img>
    <figcaption>Showing technologies used to make a work <a href="https://ewen.works/distilatex">on ewen.works</a></figcaption>
  </figure>

  <figure>
    <img src="/examples/technologies-index.png"></img>
    <figcaption>Showing all works made with Photoshop at <a href="https://ewen.works/using/photoshop">ewen.works/using/photoshop</a></figcaption>
  </figure>
</div>

## Declaration

You can define all valid technologies for your works in a central place with a `technologies.yaml` file.

Then, reference the path to that file in your `ortfodb.yaml` configuration file:

```yaml
...
  at: media/

technologies: // [!code focus]
  repository: path/to/technologies.yaml // [!code focus]

tags:
  repository: path/to/tags.yaml
```

The file itself is a list of technologies that have various properties.

### Example

```yaml
# yaml-language-server: $schema=https://ortfo.org/technologies.schema.json

- slug: go
  name: Go
  by: Google
  files: ["*.go"]
  learn more at: https://go.dev
  description: |
    An straightforward low-level open source programming language supported by Google featuring built-in concurrency and a robust standard library

- slug: vue
  name: Vue
  aliases: ["vuejs"]
  files: ["*.vue"]
  autodetect: ["vue in package.json"]
  learn more at: https://vuejs.org
  description: |
    The progressive JavaScript framework
```

::: tip What's that `# yaml-language-server` thing?
See [Declaring JSON Schemas on files](/db/json-schemas.md#on-files-directly)
:::

### Properties

<JSONSchema :schema :headings="4" type="Technology" />

## Usage

In your work's description file, refer to technologies names by their `slug`, `name` or any of the `aliases`:

```md
---
wip: true
tags: [automation, cli]
made with: [go, vue] // [!code focus]
---

# ortfo
```

## Enforcing correct technologies <Badge type=tip text=Planned />

In the future, ortfo/db might enforce that all technologies used in your works are defined in the `technologies.yaml` file.

This would prevent typos.
