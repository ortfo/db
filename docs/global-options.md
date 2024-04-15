---
editLink: false
---

# ortfodb

Manage your portfolio's database

## Synopsis

Manage your portfolio's database — See https://github.com/ortfo/db for more information.

## Examples

```
  $ ortfodb --config .ortfodb.yaml build database.json
  $ ortfodb add my-project
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &dash;&dash;config | string | config file path | ortfodb.yaml
| -h | &dash;&dash;help | | help for ortfodb 
| | &dash;&dash;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb add](add.md)	 - Add a new project to your portfolio
* [ortfodb build](build.md)	 - Build the database
* [ortfodb exporters](exporters.md)	 - Commands related to ortfo/db exporters
* [ortfodb replicate](replicate.md)	 - Replicate a database directory from a built database file.
* [ortfodb schemas](schemas.md)	 - Output JSON schemas for ortfodb's various resources

