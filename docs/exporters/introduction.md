---
next:
  text: Uploading with exporters
---

# Exporters

ortfo/db includes a plugin-like system that allows you to run shell commands at various stages of the build process.

## Usage example

Running a build with the [copy](/db/exporters/misc.md#copy), [ssh](/db/exporters/uploading.md#ssh) and [cloud](/db/exporters/uploading.md#cloud) exporters enabled


```ansi{11-99}
[1m[2m$[0m ortfodb build[0m
[1m[35m          Using[0m[0m exporter [1mcopy:[0m [2mcopy the output database file to one or more locations[0m
[1m[35m          Using[0m[0m exporter [1mssh:[0m [2mupload the database to an SSH server using scp or rsync.[0m
[1m[35m          Using[0m[0m exporter [1mcloud:[0m [2mexport the database to a cloud storage service using rclone.[0m
[1m[2m        Reusing[0m smooth-cursorify[0m
[1m[2m        Reusing[0m spotify-playlist-covers[0m
[1m[2m        Reusing[0m subfeed-for-spotify[0m
[1m[2m        Reusing[0m trigonometry-synth[0m
[1m[2m        Reusing[0m 雨と雪[0m
[1m[32m       Finished[0m[0m compiling to database.json in 0s
[1m[36m        Copying[0m[0m database to [1m~/projects/database.json, ~/projects/ortfo/website/public/example-database.json[0m[0m
[1m[34m      Uploading[0m[0m database.json to [1mewen@ewen.works:~/www/media.ewen.works/works.json [2musing rsync  [0m[0m
[1m[34m              >[0m[0m 1.181.355 100%  187,11MB/s    0:00:00 (xfr#1, to-chk=0/1)
[1m[34m      Uploading[0m[0m database.json to googledrive:projects/database.json, [2mwith rclone[0m
[1m[34m              >[0m[0m Transferring:
[1m[34m              >[0m[0m  *  database.json:100% /1.127Mi, 1.126Mi/s, 0s
[1m[34m              >[0m[0m Transferred:            1 / 1, 100%
[1m[34m              >[0m[0m Elapsed time:         3.0s
```



## Usage

To enable and configure exporters, add an `exporters` field to your `ortfodb.yaml` configuration file:

```yaml
...
  repository: path/to/tags.yaml

exporters:  // [!code focus]
  (name of the exporter):  // [!code focus]
    exporter-specific configuration...  // [!code focus]
    # available to all exporters.  // [!code focus]
    # shows more logs about the exporter, useful for debugging  // [!code focus]
    verbose: true  // [!code focus]
```

### Example

Here's an example configuration for the [FTP exporter](./uploading.md#sftp)

```yaml
exporters:
  ftp:
    host: example.com
    user: username
    password: password
    path: /path/to/remote/folder
```

## Built-in exporters

::: warning
For now, most of the built-in exporters do not work on Windows
:::

<!-- ::: tip New in v1.4.0 (not out yet)

Run [`ortfodb exporters list`](/db/commands/exporters-list) to get the list of all built-in exporters.

Run [`ortfodb exporters doc <name>`](/db/commands/exporters-doc) to get help on a specific exporter.

::: -->

### Uploading

Most built-in exporters provide functionnality to upload the generated database.json file in various ways. See [Uploading with exporters](/db/exporters/uploading.md)

### Static site generators

ortfo/db provides exporters to more conviently build your portfolio using some of the most popular static site generators. See [Static site generators](./static-site-generators)

### Exporting to other formats

ortfo/db ships with a rudimentary [SQL exporter](./formats.md#sql) to fill up a "real" database

### Other things

See [Other exporters](./misc.md).

## Running your own commands

If you don't want to write (and maybe publish) your own exporter, you can directly declare one in your configuration file by using the [Custom exporter](./custom.md)

## Creating an exporter

See [the development guide](./development.md) for more information on how to create your own exporter.
