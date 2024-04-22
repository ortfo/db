---
outline: [2, 3]
---

<script setup>
  import { onMounted } from 'vue'
  import { data } from './client-libraries.data.js'
  import { getRustVersion, getGemVersion, getGoVersion } from './client-libraries-versions.js'

  let versions = data.versions


  onMounted(async () => {
    // try to get fresh data on client-side
    versions.rust = await getRustVersion().catch(() => versions.rust)
    versions.ruby = await getGemVersion().catch(() => versions.ruby)
    versions.go = await getGoVersion().catch(() => versions.go)
  })
</script>

# Client libraries <Badge type=warning text=beta />

<div style="display: flex; justify-content: center; gap: 1em; margin: 2rem 0; flex-wrap: wrap; align-items: center">

<a href="https://pypi.org/project/ortfodb/">
  <img src="https://img.shields.io/pypi/v/ortfodb" alt="PyPI" />
</a>

<a href="https://www.npmjs.com/package/@ortfo/db">
  <img src="https://img.shields.io/npm/v/@ortfo/db" alt="npm" />
</a>

<a href="https://rubygems.org/gems/ortfodb">
  <img src="https://img.shields.io/gem/v/ortfodb" alt="RubyGems" />
</a>

<a href="https://packagist.org/packages/ortfo/db">
  <img src="https://img.shields.io/packagist/v/ortfo/db" alt="Packagist" />
</a>

<a href="https://crates.io/crates/ortfodb">
  <img src="https://img.shields.io/crates/v/ortfodb" alt="Crates.io" />
</a>

<a href="https://pkg.go.dev/github.com/ortfo/db">
  <img :src="`https://img.shields.io/badge/Go-${ versions.go }-blue`" alt="Go" />
</a>

</div>

ortfo/db exports its type definitions for use in various programming languages. This is mainly useful to access data from the exported database.json file in a type-safe way.

It's like an [ORM](https://en.wikipedia.org/wiki/Object-relational_mapping), but for your JSON database!

## Available clients

### Python

#### Installation

Install `ortfodb` [on pypi.org](https://pypi.org/project/ortfodb/):

::: code-group

```bash [with pip]
pip install ortfodb
```

```bash [with Poetry]
poetry add ortfodb
```

:::

#### Usage

```python
from ortfodb import database_from_dict
import json

with open("path/to/the/database.json") as f:
    database = database_from_dict(json.load(f))
```

### TypeScript (and JavaScript)

#### Installation

Install `@ortfo/db` [on npmjs.com](https://www.npmjs.com/package/@ortfo/db):

::: code-group

```bash [with npm]
npm install @ortfo/db
```

```bash [with yarn]
yarn add @ortfo/db
```

```bash [with pnpm]
pnpm add @ortfo/db
```

```bash [with bun]
bun add ortfodb
```

:::

#### Usage

```typescript
import { Database } from "@ortfo/db";

const database = Database.toDatabase(jsonStringOfTheDatabase);
```

### Ruby

#### Installation

Install `ortfodb` [on rubygems.org](https://rubygems.org/gems/ortfodb):

::: code-group

```ruby-vue [Gemfile]
gem 'ortfodb', '~> {{ versions.gem }}'
```

```ruby-vue{4} [my_gem.gemspec]
Gem::Specification.new do |spec|
  spec.name = 'my_gem'
  ...
  spec.add_dependency 'ortfodb', '~> {{ versions.gem }}'
end
```

:::

#### Usage

```ruby
require 'ortfodb/database'

database = Ortfodb::Database.from_json! json_string
```

### Rust

#### Installation

Install `ortfodb` [on crates.io](https://crates.io/crates/ortfodb):

::: code-group

```bash [with cargo]
cargo add ortfodb
```

```toml-vue [Cargo.toml]
[dependencies]
ortfodb = "{{ versions.rust }}"
```

:::

#### Usage

```rust
use ortfodb::Database;

let database: Database = serde_json::from_str(&json_string).unwrap();
```

### PHP

See https://github.com/ortfo/db/tree/main/packages/php

::: warning
It's not on Packagist yet because packagist expects a `composer.json` file at the root of the repository
:::

#### Installation

TODO

#### Usage

::: warning
I haven't programmed in PHP in a while, so this might be wrong. Please open an issue if you find a mistake.
:::

```php
<?php

use Ortfo\Db\Database;

$json = json_decode(file_get_contents('path/to/the/database.json'), true);
$database = Database::from($json);
```

## Using Go?

Since ortfo/db is programmed in Go, you can use the ortfo/db's Go package directly, see the [Go package documentation](https://pkg.go.dev/github.com/ortfo/db). The [`Database` type](https://pkg.go.dev/github.com/ortfo/db#Database) is probably what you're looking for.

## Client libraries versions

The versions of the package are synchronized with the versions of ortfo/db, so you can always know which version of the client library to use.

## How it's done

The client libraries are generated from the various [JSON schemas](/db/json-schemas.md) of ortfo/db, thanks to [quicktype](https://quicktype.io/). This way, they are always up-to-date!
