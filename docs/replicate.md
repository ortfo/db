---
editLink: false
---

# ortfodb replicate

Replicate a database directory from a built database file.

## Synopsis

Replicate a database from <from-filepath> to <to-filepath>. Note that <to-filepath> must be an empty directory.

Example: ortfodb replicate ./database.json ./replicated-database/

WARNING: This command is still kind-of a WIP, it works but there's minimal logging and error handling.


```
ortfodb replicate <from-filepath> <to-filepath> [flags]
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -h | &hyphen;&hyphen;help | | help for replicate 
| -n | &hyphen;&hyphen;no-verify | | Don't try to validate the built database file before replicating 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb](global-options.md)	 - Manage your portfolio's database

