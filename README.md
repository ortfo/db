# portfoliodb

****
> I'm following RDD (readme-driven development) for this project, so, until v0.1.0 is released, this document describes what the program will look like
****

A readable, easy and enjoyable way to manage portfolio databases using directories and text files.

## Installation

Pre-compiled binaries are available through [GitHub Releases](https://help.github.com/en/github/administering-a-repository/releasing-projects-on-github):

```shell
$ wget https://github.com/ewen-lbh/portfoliodb/releases/latest/portfoliodb
# Put the command in a directory that is in your PATH, so that you can use portfoliodb from anywhere, e.g.:
$ mv portfoliodb /usr/bin/portfoliodb
```

See [Compiling](#compiling) for instructions on how to compile this yourself

## Usage

See usage [here](./USAGE)

## How it works

Your database is a folder, which has one folder per work in it.
In each folder, you'll have a markdown file describing your work, and some files, which can be videos, audio files or images.

Here's an example tree:

```directory-tree
database/
├── ideaseed
│   ├── logo.png
│   └── description.md
├── phelng
│   └── description.md
├── portfolio
│   └── description.md
├── portfoliodb
│   └── description.md
└── schoolsyst
    ├── description.md
    ├── api
    │   └── description.md
    ├── presentation
    │   └── description.md
    └── webapp
        └── description.md
```

### `description.md` files

The most information is extracted from natural markdown, however, you can add some information via a YAML header.
Here's an example `description.md`

Information inside YAML headers is arbitrary, put any key and it will be added to the work's JSON object.

Some YAML keys are interpreted in a certain way though:

- `collection`: This _needs_ to refer to a valid collection id, and associates that work with the collection
- `color`: nothing special about this, but note that it overrides `extract color from`. 
  It can be an object (with keys `secondary` and `primary`) or a string 
  (which is the same as setting `primary` and setting `secondary` to null.
- `extract color from`: portfoliodb has an optional build step to extract the primary and secondary colors from a given image, 
  this sets the filename. With the build step turned on and without this, _portfoliodb_ will look for image files in the work's folder,
  and if it founds only one image, it will use it, else the build will fail.

Other information is extracted from the contents themselves:

- `name` is extracted from the document's title (`# phelng` here)
- `tags` is extracted from the first paragraph following the title, and are split on commas and newlines (`[cli, automation, program]` here)
- `links` is extracted from a list that only contains named links (`- [code source](https://github.com/ewen-lbh/phelng)` here)
- `made_with` is extracted from an unordered list following an `<h2>` named "Made with" (`## Made with`), each list item become a string and is added to `using` (`[python]` here)

A special syntax is added to easily embed video or audio files, either from files in the work's folder, or from YouTube (playlists and videos):

```markdown
>[fallback text "Optional title"](filename or youtube URL)
```

This syntax was chosen to ressemble the image's, and because the `>` symbolizes a "play" button.
It also renders not-too-bad on regular markdown, embedding a media can be sort of seen as a quotation

YouTube URLs always become video embeds and for local files, the MIME type and extension are checked to determine if it's an audio or video file.

Each description "chunk" is separaed by an horizontal rule (`----`).

The first chunk is added to `summary`, while others are added as an array of strings to `description_chunks`

```markdown
---
created: 2020-05
wip: yes
best: yes
color: FFFFFF
---

<!-- name -->

# phelng

<!-- tags -->

cli,
automation,
program

<!-- links -->

- [code source](https://github.com/ewen-lbh/phelng)

<!-- youtube -->

>[video](https://www.youtube.com/watch?v=qj2fglI1sYw)

Un script pour télécharger automatiquement des musiques, en utilisant YouTube comme source audio et Spotify comme source de métadonnées.
Pour chaque morceau, la bonne vidéo YouTube est sélectionnée en coïncidant des informations telles que la durée du morceau et celle de la vidéo,
puis est téléchargée, les métadonnées (incluant une image de la pochette) est appliquée, puis le volume sonore est normalisé.

----

La liste de morceaux à télécharger est stockée dans un fichier `.tsv`:

Au début, j'ai eu l'idée d'utiliser le format de fichier le plus simple et le plus intuitif possible : un fichier texte, chaque piste d'une ligne, au format `{artiste} - {titre}`.

Mais il y a quelques problèmes avec cette technique:

- **Les noms d'artistes ne peuvent pas contenir " - "**

  ...ce qui ne se produirait jamais de toute façon, mais avec [les titres des pistes que _Four Tet_ propose](https://open.spotify.com/album/6iFZ3Kcx8CDmcMNyKRqUwc?highlight=spotify:track:3bCs4oOGpM0KkVB78Laiqp), il ne faut jamais sous-estimer la créativité parfois affichée dans les titres).

- **On ne peut pas ajouter d'informations supplémentaires sans restreindre les titres des pistes**

  Certains artistes aiment ajouter des titres "Intro" et/ou "Outro" à leurs albums, par exemple
  Imaginez qu'un artiste ait deux albums _A_ et _B_, chacun ayant une piste d'intro nommée exactement _Intro_.
  Si vous voulez télécharger la _Intro_ de **_B_**, vous ne pouvez pas le spécifier.
  Une nouvelle syntaxe pourrait être introduite, quelque chose comme "artiste - piste [album]" mais, encore une fois, que faire si le titre de la piste contient un crochet ouvrant "[" ?

----

La solution : utiliser un caractère _tab littéral_ comme séparateur d'informations.

Et certains y ont déjà pensé, nous avons donc l'avantage d'utiliser un langage déjà existant :
le format de fichier _tsv_, ou valeurs séparées par des tabulations.
Cela signifie également que vous pouvez facilement modifier et consulter votre bibliothèque dans
n'importe quel logiciel de tableur.

La seule mise en garde pour ce cas d'utilisation avec les fichiers tsv est qu'il n'existe pas de
norme pour les commentaires. Les commentaires peuvent être utiles dans votre fichier de bibliothèque
pour "désactiver" temporairement les pistes et empêcher leur téléchargement, ou pour servir d'en-tête
au début du fichier pour vous aider à vous souvenir du format.

----

Comme nous ne voulons pas limiter les caractères que les noms d'artistes peuvent contenir, nous ne pouvons pas utiliser quelque chose comme "un commentaire" ou "un autre". Comme les deux premiers champs sont _requis_, le simple fait d'avoir une ligne qui commence par un caractère de tabulation est ignoré par _phelng_, et signifierait autrement que la colonne "artiste" est indéfinie pour cette ligne.

Ainsi, le format (jusqu'à présent*) est le suivant (avec `⭾` représentant un caractère de tabulation)

    commentaire ⭾A (ignoré par phelng)
    Artist⭾Track title⭾Album (facultatif)⭾Duration (en quelques secondes, facultatif)

*Des champs supplémentaires pourraient être ajoutés à l'avenir, _sans rupture de la compatibilité ascendante_, puisque l'ordre des champs sera _toujours conservé_.

<!-- using -->

## Made with

- python
```

## Configuration

Put this in `.portfoliodb.yml` in the root of your database:

```yaml
build steps:
  - step: extract colors
    extract:
      - primary
      - secondary
      - tertiary
    default file name: logo.png

  - step: make gifs
    # <filetitle> refers to the filename without its extension.
    file name template: <filetitle>.gif

  - step: make thumbnails
    widths: [20, 100, 500]
    input file: logo.png # If the directory has one image file and no logo.png, it uses this file instead
    # Paths are always relative to the work's database folder
    file name template: ../../static/thumbs/<id>/<width>.png


features:
  # Specified by an <ul> directly after a <h2>Made with</h2>
  made with: on
  # Extract media from the document and put it in an object:
  # { media: [type]: [{ name, url }, {name, url}, ...] }
  # with [type] one of audio, image and video
  media hoisting: off

validate:
  checks:
    # can be `off` (not checked for)
    # can be `on` (uses the default level)
    # can be a level:
    # - `fatal`: also checked when building, triggers end of build if fails
    # - `error`: prints an error message (red), makes validate command exit with 1
    # - `warn` : prints a warning message (orange), does not make validate exit with 1
    # - `info` : regular message, informative
    # these are the default values
    schema compliance: fatal
    work folder uniqueness: fatal
    work folder safeness: error
    yaml header: error
    title presence: error
    title uniqueness: error
    tags presence: warn
    tags knowledge: error
    working media: warn
    working urls: off
```

PRO TIP: You can use the provided `.portfoliodb.yml.schema.json` to validate your YAML file
with this JSONSchema

## Extra markdown features

Except for the `>[text](video/audio URL/filepath)` feature, the markdown also supports a number of non-standard features:

- all of what GFM supports (except autolinking of issues and commit hashes, ofc)
- Abbreviations: `*[YAML]: Yet Another Markup Language`
- Definition lists: `- key: value` or the more standard, [PHP-markdown-extra-style](https://michelf.ca/projects/php-markdown/extra/#def-list)
- Admonitions: `!!! type "Optional title"`, see [this documentation](https://python-markdown.github.io/extensions/admonition/)
- Footnotes: `footnote reference[^1]` and then `[^1]: footnote content`
- Markdown in HTML: [See documentation here](https://python-markdown.github.io/extensions/md_in_html/)
- (off by default) New-line-to-line-break: Transforms line breaks in markdown into `<br>`s, see [the documentation](https://python-markdown.github.io/extensions/nl2br/)
- Smarty pants: typographic replacements (not replaced inside code):
  - `--` to –
  - `---` to —
  - `->` to →
  - `<-` to ←
  - `...` to …
  - `<<` to «
  - `>>` to »
- (off by default) Anchored headings: Each headings is assigned an id to reference in the URL with `example.com#heading`

### Configuring markdown

The extra features discussed just above are all available or disable, using the module name:

_.portfoliodb.yml_
```yaml
markdown:
  abbreviations: on
  definition lists: on
  admonitions: off
  footnotes: on
  markdown in html: on
  new-line-to-line-break: on
  smarty pants: off
  anchored headings:
  # you can also use an object form to pass in config options
    enabled: yes
    format: <content> # default value
  custom syntaxes:
    # this is just an example, not an actual implementation of the video/audio embed feature
    - from: '>\[(<fallback>[^\]]+)\]\((<source>.+)\)'
      to: <video src="${source}">${fallback}</video>
```

# Compiling

1. Clone the repository: `git clone https://github.com/ewen-lbh/portfoliodb`
2. `cd` into it: `cd portfoliodb`
3. `make` the binary: `make`
4. Install it (this just copies the file to `/usr/bin/`): `make install`
