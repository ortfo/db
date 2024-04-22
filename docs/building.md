---
next: false
---

# Running ortfo/db

## Building your database

Make sure that ortfodb in [installed](/db/getting-started.md#installation).

Open up a terminal, and run the following:

::: code-group

```sh [Regular mode]
ortfodb build database.json
```

```sh [Scattered mode]
ortfodb --scattered build database.json
```

:::

You should get the following output:

```ansi
[1m[33m        Writing[0m[0m default configuration file at ortfodb.yaml
[1m[33m        Warning[0m[0m default configuration assumes that your projects
                live in /mnt/datacore/projects/ortfo/website.
                Change this with [1mprojects at[0m in the generated configuration file[0m
[1m[34m           Info[0m[0m No configuration file found. The default configuration was used.
[1m[32m       Finished[0m[0m compiling to database.json
```

Subsequent runs of this command will be faster as long as you re-use the output file: ortfo/db will only recompile the projects that have changed since the last run.

Notice the warning. If you ran the previous command from the directory that contains all of your projects, you should be fine. But if you ran it from somwhere else, you'll probably want to change that `projects at` setting it's talking about to point it to where your projects are.


## Tweaking the configuration file

Open up `ortfodb.yaml` in your favorite text editor.

It should look a little like this:

```yaml
extract colors:
    enabled: true
    extract: []
    default files: []
make gifs:
    enabled: false
    file name template: ""
make thumbnails:
    enabled: true
    sizes:
        - 100
        - 400
        - 600
        - 1200
    input file: ""
    file name template: <work id>/<block id>@<size>.webp
build metadata file: .lastbuild.yaml
media:
    at: media/
scattered mode folder: .ortfo
tags:
    repository: ""
technologies:
    repository: ""
projects at: /home/you
```

The config file that ortfo/db generated for you has mostly sensible defaults. The most important things to check are:

`projects at`
: The path to the directory that contains all of your projects

`media.at`
: Where to copy all the media files you reference in your description.md files, as well as the [generated thumbnails](/db/thumbnails)

Most other options relate to certain features, you'll find documentation about them in the pages relating to the features themselves.

::: tip TODO
TODO: Document the whole config file in one place
:::

## What now?

Congrats, you've setup ortfo/db!

Next steps:

What to know all the data contained in your database?
: Check out [the database format](/db/database-format.md)

Want to learn more about the different features ortfo/db has to offer?
: Check out the [features page](/db/features.md)

Wanna integrate ortfo/db with a static site generator?
: See the [Static site generator exporters](/db/exporters/static-site-generators.md)

Curious about how ortfo/db is used to actually make a portfolio site?
: Check out [the repository](https://github.com/ewen-lbh/portfolio) for [my own portfolio](https://ewen.works), which is (of course) built with ortfo/db.
: You can also see how [net7](https://github.com/inp-net) uses ortfo/db to keep their [projects page](https://net7.dev/realisation.html) up to date: see [the repository](https://git.inpt.fr/net7/website/-/tree/master?ref_type=heads) (warning: french ahead)
