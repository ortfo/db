build:
	#!/usr/bin/env bash
	set -euxo pipefail
	cd cmd
	go mod tidy
	go build
	mv cmd ../ortfodb

install:
	just build
	cp ortfodb ~/.local/bin/ortfodb
	chmod +x ~/.local/bin/ortfodb

docs:
	mkdir -p docs manpages
	./ortfodb makedocs

render-demo-gif:
	#!/usr/bin/env bash
	set -euxo pipefail
	cd ~/projects/portfolio
	just db
	jq 'delpaths([[".ortfo", ".centraverse", ".onset"]])' < database.json | sponge database.json
	vhs ~/projects/ortfo/db/demo.tape -o ~/projects/ortfo/db/demo.gif

prepare-release $VERSION:
	./tools/update_meta_go.py $VERSION
	git add meta.go
	git commit -m "temp commit for release $VERSION"
	git tag v$VERSION
	# only build & create archives, publishing & packaging is done later.
	# i can't use goreleaser --prepare because i don't have like 15 fking dollars to spend each month on developer tooling lmao
	goreleaser release --clean \
		--skip announce \
		--skip aur \
		--skip chocolatey \
		--skip docker \
		--skip homebrew \
		--skip ko \
		--skip nfpm \
		--skip nix \
		--skip publish \
		--skip sbom \
		--skip scoop \
		--skip sign \
		--skip snapcraft \
		--skip validate \
		--skip winget
	git reset --soft HEAD^
	git tag -d v$VERSION
	just build
	just docs
	./tools/generate_schemas.py
	./tools/build_readme.py
	just build-client-libraries $VERSION

release name='${version}':
	GITHUB_TOKEN=$(rbw get 'GitHub VSCode PAT') release-it --github.releaseName={{quote(name)}}

publish:
	just publish-client-libraries
	just package

publish-client-libraries:
	cd packages/python; poetry publish
	cd packages/typescript; npm publish
	cd packages/rust; cargo publish
	cd packages/ruby; gem push ortfodb-*.gem; rm ortfodb-*.gem
	# TODO: PHP. Packagist wants the repo all to itself, so I have to create a new repo, copy the code in it; etc.

package:
	wget https://ortfo.org/android-chrome-512x512.png -O icon.png
	printf 'FROM scratch\nENTRYPOINT ["ortfodb"]\nCOPY ortfodb /\n' > Dockerfile
	GITHUB_TOKEN=$(rbw get 'GitHub VSCode PAT') AUR_KEY=~/.ssh/id_arch_aur goreleaser --verbose release || rm icon.png Dockerfile; exit 1
	rm icon.png Dockerfile

build-client-libraries version:
	just build-typescript {{version}}
	just build-python {{version}}
	just build-rust {{version}}
	just build-ruby {{version}}
	just build-php {{version}}

build-php version:
	#!/usr/bin/env bash
	set -euxo pipefail
	for schema in schemas/*.schema.json; do
		pascal_case=$(basename $schema .schema.json | sed -re 's/(^|-)([a-z])/\U\2/g')
		quicktype --src-lang schema -l php $schema -o packages/php/src/$pascal_case.php
		sed -i 's/<?php/<?php\n\nnamespace Ortfo\\Db;/g' packages/php/src/$pascal_case.php
	done
	cd packages/php
	composer install

build-ruby version:
	#!/usr/bin/env bash
	set -euxo pipefail
	for schema in schemas/*.schema.json; do
		quicktype --src-lang schema -l ruby $schema -o packages/ruby/lib/ortfodb/$(basename $schema .schema.json).rb --namespace Ortfodb
	done
	cd packages/ruby
	printf 'module Ortfodb\n\tVERSION = "%s"\nend\n' {{ version }} > lib/ortfodb/version.rb
	gem build ortfodb.gemspec

build-typescript version:
	#!/usr/bin/env bash
	set -euxo pipefail
	for schema in schemas/*.schema.json; do
		quicktype --src-lang schema -l typescript $schema -o packages/typescript/src/$(basename $schema .schema.json).ts
	done
	cd packages/typescript
	jq '.version = "{{ version }}"' < package.json | sponge package.json
	rm -r dist
	npm run build

build-python version:
	#!/usr/bin/env bash
	set -euxo pipefail
	for schema in schemas/*.schema.json; do
		quicktype --src-lang schema -l python $schema -o packages/python/ortfodb/$(basename $schema .schema.json).py
	done
	cd packages/python
	poetry version {{ version }}
	poetry build

build-rust version:
	#!/usr/bin/env bash
	set -euxo pipefail
	for schema in schemas/*.schema.json; do
		quicktype --src-lang schema -l rust --visibility public $schema -o packages/rust/src/$(basename $schema .schema.json).rs
		sed -i 's/generated_module/ortfodb/g' packages/rust/src/$(basename $schema .schema.json).rs
	done
	cd packages/rust
	tomlq -ti '.package.version = "{{ version }}"' Cargo.toml
	cargo build
