---
outline: [1, 2]
---

# Static site generators

ortfo/db provides exporters to more conviently build your portfolio using some of the most popular static site generators.

::: tip
While ortfo/mk is [not ready at the moment](/guide/what-is-ortfo.md#the-plan), a static site generator that's made for ortfo/db is planned. It will be able to read the database directly, and will be able to generate a portfolio website with minimal configuration.
:::

## Hugo

[Hugo](https://gohugo.io) is a static site generator written in Go. It is known for its speed and flexibility. The `hugo` exporter exports each work as a data file for use in Hugo's Data templates


### Usage

Add the `hugo` exporter to your `ortfodb.yaml` configuration file:

```yaml
exporters:
  hugo:
    # will write JSON for individual works at data/<work id>.json
    in: data/works
    # how to call the index file, relative to the first directory of the `in`
    index: index.json
```

## 11ty <Badge type=warning text=Experimental />

[11ty](https://www.11ty.dev) is a simpler static site generator written in JavaScript. It is known for its simplicity and flexibility. The exporter exports data template directory files for use in 11ty.js (Work in progress, never tested but should work fine)

### Usage

Add the `11ty` exporter to your `ortfodb.yaml` configuration file:

```yaml

exporters:
  11ty:
    # will write JSON for individual works at projects/<work id>.11tydata.json
    in: projects/
```

## Others

[Creating your own](./development.md) exporter is really easy, and you might not even need one. Since ortfo/db outputs a JSON file, you can use it with any static site generator that can read JSON files (which I hope is most of them).

These exporters are at most a convenience.

If you do make one, please add it to this page! The "Edit this page" button is just below ;)
