---
editLink: false
---

# ortfodb schemas

Output JSON schemas for ortfodb's various resources

## Synopsis

Don't pass any resource to get the list of available resources

Output the JSON schema for:
- configuration: the configuration file (.ortfodb.yaml)
- database: the output database file
- tags: the tags repository file (tags.yaml)
- technologies: the technologies repository file (technologies.yaml)
- exporter: the manifest file for an exporter


```
ortfodb schemas [resource] [flags]
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -h | &hyphen;&hyphen;help | | help for schemas 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb](global-options.md)	 - Manage your portfolio's database

