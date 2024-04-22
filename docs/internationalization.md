# Internationalization

To share your works with the world, and with your future job, you might want to translate your works' descriptions to multiple languages. ortfo/db has _first-class_ support for this by allowing you to define your works' descriptions in multiple languages.

## Usage example

<div class="side-by-side">
  <figure>
    <img src="/examples/internationalization-en.png"></img>
    <figcaption>A work in English <a href="https://en.ewen.works/spotify-playlist-covers">on en.ewen.works</a></figcaption>
  </figure>

  <figure>
    <img src="/examples/internationalization-fr.png"></img>
    <figcaption>The same work in French <a href="https://fr.ewen.works/spotify-playlist-covers">at fr.ewen.works</a></figcaption>
  </figure>
</div>

## Language markers

Simply add a _language marker_ in your description file on its own line. Everything that you write after this marker, and before the next marker, is considered to be written in that language.

The markers look like this:

```md
:: language code
```

Note that this “language code” could technically be anything you want, but it's recommended (and might in the future be required to be) a [BCP 47](https://en.wikipedia.org/wiki/IETF_language_tag#List_of_common_primary_language_subtags) language code.

### Example

```md
---
wip: yes
---

# ortfo

:: en

A simple way to manage _lots_ of projects for a portfolio website

:: fr

Une manière simple de gérer _beaucoup_ de projets pour un site web de portfolio

:: ja

私の日本語は悪いですから、知らない wwww
```

## Untranslated content

Content that isn't translated is considered to have the special language code `default`.

All content before any language markers is considered untranslated.

This is particularly useful for the title of the work, which is usually the same in all languages.

## In `database.json`

The database's [Content](/db/database-format.md#content) will be an object mapping every language code used in the description file to the content blocks of the work.

## `layout` considerations

Since the [Layout](/db/layouts.md) is shared by all languages, you must have the same number of paragraph, links and media blocks in every language.
