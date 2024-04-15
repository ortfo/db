---
editLink: false
---

# ortfodb build

Build the database

## Synopsis

Scan in the projects directory for folders with description.md files (and potential media files) and compile the whole database into a JSON file at <to-filepath>.

If <to-filepath> is "-", the output will be written to stdout.

If [include-works] is provided, only works that match the pattern will be included in the database.


```
ortfodb build <to-filepath> [include-works] [flags]
```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -e | &dash;&dash;exporters | stringArray | Exporters to enable. If not provided, all the exporters configured in the configuration file will be enabled. 
| -h | &dash;&dash;help | | help for build 
| -m | &dash;&dash;minified | | Output a minifed JSON file 
| | &dash;&dash;no-cache | | Disable usage of previous database build as cache for this build (used for media analysis among other things). 
| -q | &dash;&dash;silent | | Do not write to stdout 
| | &dash;&dash;workers | int | Use <count> workers to build the database. Defaults to the number of CPU cores. | 12
| | &dash;&dash;write-progress | string | Write progress information to a file. See https://pkg.go.dev/github.com/ortfo/db#ProgressInfoEvent for more information. 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &dash;&dash;config | string | config file path | ortfodb.yaml
| | &dash;&dash;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb](global-options.md)	 - Manage your portfolio's database

