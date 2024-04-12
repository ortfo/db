build:
	#!/usr/bin/env bash
	set -euxo pipefail
	cd cmd/ortfodb
	go mod tidy
	go build
	mv ortfodb ../../

install:
	just build
	cp ortfodb ~/.local/bin/ortfodb
	chmod +x ~/.local/bin/ortfodb

prepare-release $VERSION:
	./tools/update_meta_go.py $VERSION
	just build
	./tools/generate_schemas.py
	./tools/build_readme.py

release name="":
	GITHUB_TOKEN=$(rbw get 'GitHub VSCode PAT') release-it --github.releaseName="{{name}}"
