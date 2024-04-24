# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- field abbreviations to paragraph blocks that maps abbreviations to their HTML definitions
- `json` tags to Tag and Technology. could be useful to serialize to _JSON_ (and not YAML) the tags and or techs repository files, without having weird keys in the resulting JSON
- errors when encountering duplicate block IDs, empty blocks and other edge cases in descriptions
- support for `null` values in layouts. this gets rendered as a special layout cell "ghost" (see `ortfodb.EmptyLayoutCell` for the value). This is useful to add empty spaces in a grid of images, for example
- lsp subcommand for LSP support

### Changed

- building the database is now significantly faster!
- renamed AnalyzedWork type to Work

### Removed

- build metadata file (`.lastbuild.yaml`)

### Fixed

- weird "no non-transparent pixels found" error when trying to extract colors from .gif files.
- progress bar not finishing when skipping works with an include argument to build

## [1.5.0] - 2024-04-20

### Added

- a crystal client library (work in progress)
- variables to write logs to a file

### Changed

- exposed all types in Rust and Python client libraries
- move ReleaseBuildLock out of RunContext
- allow user to only specify some colors: unspecified colors will get their extracted color
- the --write-progress progress file is now removed when the build is done
- add command: add link block to source code if in a git repo

### Fixed

- json schemas for tags.yaml and technologies.yaml: all fields were wrongly marked as required
- json schemas configuration.yaml: all fields were wrongly marked as required
- layout was not normalized in output database

## [1.4.1] - 2024-04-17

### Added

- warning about the `projects at` value when using default config file
- support for NO_COLOR and FORCE_COLOR env vars

### Changed

- move `Log*` functions out of `RunContext`

### Fixed

- the default configuration included the media directory in the media.at value, making ortfodb write media files in `media/media/`

## [1.4.0] - 2024-04-16

### Added

- exporters: sub-commands list and doc
- new exporter "webhook" to send a POST to some URL with the built database

### Changed

- built-in exporters are now all embedded in the binary

## [1.3.0] - 2024-04-16

### Added

- exporters init command: bootstrap a new exporter manifest file
- exporters: [sprig](https://masterminds.github.io/sprig/) functions are now available in templates, along with shell-escape strings
- env variables ORTFO_DEBUG and ORTFODB_DEBUG can now be used as alternatives to DEBUG to enable debug mode. as with DEBUG, the value must be "1" to enable debug mode
- cloud exporter: config "name" to rename the uploaded file. defaults to the local file name

### Fixed

- tags.repository and technologies.repository now handle expansion of ~ and ~user

## [1.2.0] - 2024-04-14

### Added

- (S)FTP exporter
- Git exporter: clones a repo, adds and commits the database json file and pushes
- Cloud exporter: uses rclone to upload the database.json file to many cloud services
- Requires key in exporter manifests to specify programs required to run the exporter
- localize exporter: export the database as a single-language database for every language in the original database

### Changed

- building now shows exporters that are activated along with their description

## [1.1.0] - 2024-04-14

### Added

- exporters: run custom shell commands before and after the build, and/or after each work is built.
- SQL exporter: a rudimentary SQL exporter, written in the Go code directly
- SSH exporter: a rudimentary SSH exporter that uploads the built database somewhere via ssh. written as a normal YAML exporter, see exporters/ssh.yaml

## [1.0.0] - 2024-04-13

### Added

- `completion` command to install completions for your shell!

### Changed

- restructed the command line interface to be more "normal"

### Fixed

- the add command would add the `databaseMeta` key in the generated frontmatter

## [0.3.2] - 2024-04-13

### Fixed

- replace meta-work in database with a databaseMetadata field on all works' metadata. client libraries should be able to generate properly from the resulting, simpler JSON schema.

## [0.3.1] - 2024-04-13

### Fixed

- invalid json schema for database

## [0.3.0] - 2024-04-12

### Changed

- add command: detect summary and title from README.md, and finish date from git history
- updated --help information for --write-progress' file format

## [0.2.0] - 2024-04-12

### Added

- Initial release

[Unreleased]: https://github.com/ortfo/db/compare/v1.5.0...HEAD
[1.5.0]: https://github.com/ortfo/db/compare/v1.4.1...v1.5.0
[1.4.1]: https://github.com/ortfo/db/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/ortfo/db/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/ortfo/db/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/ortfo/db/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ortfo/db/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ortfo/db/compare/v0.3.2...v1.0.0
[0.3.2]: https://github.com/ortfo/db/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/ortfo/db/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/ortfo/db/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ortfo/db/releases/tag/v0.2.0

[//]: # (C3-2-DKAC:GGH:Rortfo/db:Tv{t})
