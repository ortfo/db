---
editLink: false
---

# ortfodb add

Add a new project to your portfolio

## Synopsis

Create a new project in the appropriate folder. ID is the work's slug.

```
ortfodb add <id> [flags]
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -h | &hyphen;&hyphen;help | | help for add 
| | &hyphen;&hyphen;overwrite | | Overwrite the description.md file if it already exists 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb](global-options.md)	 - Manage your portfolio's database

