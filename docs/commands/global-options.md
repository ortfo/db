---
editLink: false
---

# ortfodb

Manage your portfolio's database

## Synopsis

Manage your portfolio's database — See https://github.com/ortfo/db for more information.

## Examples

```ansi
[1m[2m$[0m [1mortfodb[0m [36m--config[0m [32m.ortfodb.yaml[0m [34mbuild[0m [32mdatabase.json[0m
[1m[2m$[0m [1mortfodb[0m [34madd[0m [32mmy-project[0m[0m
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| -h | &hyphen;&hyphen;help | | help for ortfodb 
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb add](add.md)	 - Add a new project to your portfolio
* [ortfodb build](build.md)	 - Build the database
* [ortfodb exporters](exporters.md)	 - Commands related to ortfo/db exporters
* [ortfodb lsp](lsp.md)	 - Start a Language Server Protocol server for ortfo
* [ortfodb replicate](replicate.md)	 - Replicate a database directory from a built database file.
* [ortfodb schemas](schemas.md)	 - Output JSON schemas for ortfodb's various resources

