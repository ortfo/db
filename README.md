# ortfo/db

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

![](./demo.gif)

## Installation

Pre-compiled binaries are available in the [Releases](https://github.com/ortfo/db/releases) page.

See the [documentation on installation](https://ortfo.org/db/getting-started#installation) for more information


## Usage

See [documentation](https://ortfo.org/db)

```docopt
Manage your portfolio's database — See https://github.com/ortfo/db for more information.

Usage:
  ortfodb [command]

Examples:
  $ ortfodb --config .ortfodb.yaml build database.json
  $ ortfodb add my-project

Available Commands:
  add         Add a new project to your portfolio
  build       Build the database
  completion  Generate the autocompletion script for the specified shell
  exporters   Commands related to ortfo/db exporters
  help        Help about any command
  replicate   Replicate a database directory from a built database file.
  schemas     Output JSON schemas for ortfodb's various resources

Flags:
  -c, --config string   config file path (default "ortfodb.yaml")
  -h, --help            help for ortfodb
      --scattered       Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/
  -v, --version         version for ortfodb

Use "ortfodb [command] --help" for more information about a command.

```
