---
editLink: false
---

# ortfodb exporters doc

Get help for a specific exporter

```
ortfodb exporters doc <name> [flags]
```

## Examples

```ansi
$ ortfodb exporters help localize

[1m[34mlocalize  [0m  Export separately the database as a single database for each language. The
`content` field of each work is localized, meaning it's not an object mapping
languages to localized content, but the content directly, in the language.
[0m            [1m[36mOptions[0m:
[0m            [1m[2m•[0m [34mfilename_template[0m
[0m

To add [1mlocalize[0m to your project, add the following to [36myour ortfodb config file[0m:

[0m  [1m[2m[31mexporters:
[0m    [1m[31mlocalize:[0m [2m# <- add this alongside your potential other exporters
[0m      [1m[31mfilename_template:[0m [32m""[0m
[0m
Feel free to change these configuration values. Check out the exporter's documentation to learn more about what they do.

```

## Options

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -h | &hyphen;&hyphen;help | | help for doc 

## Options inherited from parent commands

| Shorthand | Flag | Argument | Description | Default value |
| --- | --- | --- | --- | --- |
| -c | &hyphen;&hyphen;config | string | config file path | ortfodb.yaml
| | &hyphen;&hyphen;scattered | | Operate in scattered mode. In scattered mode, the description.md files are searched inside `.ortfo' folders in every folder of the database directory, instead of directly in the database directory's folders. See https://github.com/ortfo/ 

## See also

* [ortfodb exporters](exporters.md)	 - Commands related to ortfo/db exporters

