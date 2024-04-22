<script setup>
  import { onMounted } from 'vue'
  import { getGoVersion } from './client-libraries-versions.js'
  import { data } from './json-schemas.data.js'
  let version = data.version
  onMounted(async () => {
    version = await getGoVersion().catch(() => "v1.2.0")
  })
</script>

# JSON Schemas

ortfo/db exports [JSON Schemas](https://json-schema.org/) for the different data files it uses.

These schemas serve as both validation when running the program, and as a way to provide a nice [auto-complete experience in your editor](#using-it-in-your-editor), provided that it supports JSON schemas.

::: tip Client libraries

Want type definitions in your code? Check out the [client libraries](/db/client-libraries.md) <Badge type=warning text=beta /> for your language.

:::

## Getting the schemas

### Locally

See [`ortfodb schemas`](/db/commands/schemas.md)

### Over the network

The schemas are all available on ortfo/db's repository in the `schemas/` directory, and are re-exported to the ortfo.org domain for easier access.

[ortfo.org/configuration.schema.json](https://ortfo.org/configuration.schema.json)
: The schema for the `ortfodb.yaml` [configuration file](/db/configuration.md)

[ortfo.org/database.schema.json](https://ortfo.org/database.schema.json)
: The schema for the `database.json` file (see [Database format](/db/database-format.md))

[ortfo.org/tags.schema.json](https://ortfo.org/tags.schema.json)
: The schema for the [tags repository](/db/tags.md)

[ortfo.org/technologies.schema.json](https://ortfo.org/technologies.schema.json)
: The schema for the [technologies repository](/db/technologies.md)

#### Version pining

Instead of getting the latest version, you can get a specific version by specifying it in the URL before the file name:

<center style="margin-bottom: 2em">
<span style="color: gray">https://</span>ortfo.org/<strong><span style="color: var(--vp-c-brand-1)">{{ version }}</span></strong>/<em>resource name</em>.schema.json
</center>

## Using it in your editor

A motivating example, for the [tags repository](/db/tags.md) file:

```yaml
# yaml-language-server: $schema=https://ortfo.org/tags.schema.json

- singular: website
  plural: websites
  aliases: site # expected array of strings, got string // [!code error]
```

RedHat develops a [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) for YAML that [includes JSON Schema support](https://github.com/redhat-developer/yaml-language-server?tab=readme-ov-file#language-server-settings), so most IDEs that support LSPs should be able to support this feature.

### On files directly

You can add a magic comment at the top of your file to associate it with a JSON schema [[docs]](https://github.com/redhat-developer/yaml-language-server?tab=readme-ov-file#using-inlined-schema):

```yaml{1}
# yaml-language-server: $schema=https://ortfo.org/configuration.schema.json

make thumbnails:
  enabled: true
  ...
```

### Associating by filename

You can associate [glob patterns](<https://en.wikipedia.org/wiki/Glob_(programming)>) of YAML filenames with JSON schemas in your editor settings, [see the documentation](https://github.com/redhat-developer/yaml-language-server?tab=readme-ov-file#associating-a-schema-to-a-glob-pattern-via-yamlschemas)

## IDE support

VSCode
: [RedHat YAML Extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

Neovim
: [coc-yaml](https://github.com/neoclide/coc-yaml), a [CoC](https://github.com/neoclide/coc.nvim) plugin
: Any LSP client plugin, such as [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) should also do the trick

_Other editors_
: Please contribute to the docs (see "Edit this page" just below)
