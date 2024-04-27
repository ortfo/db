---
editLink: false
---

# ortfodb lsp

Start a Language Server Protocol server for ortfo

```
ortfodb lsp [flags]
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -h | &hyphen;&hyphen;help | | help for lsp 
| | &hyphen;&hyphen;stdio | | Used for compatibility with VSCode. Ignored (the server is always started in stdio mode) 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb](global-options.md)	 - Manage your portfolio's database

