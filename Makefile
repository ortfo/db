.PHONY: json_schemas.go examples

build:
	cd cmd/ortfodb; \
		go mod tidy; \
		go build;

install:
	cp cmd/ortfodb/ortfodb ~/.local/bin/ortfodb
	chmod +x ~/.local/bin/ortfodb

readme:
	./tools/build_readme.py

json_schemas.go:
	./tools/build_configuration_schema_go.py

dev:
	filewatcher -I "{**.go,configuration.schema.json,_README.md}" -x configuration_schema.go "make configuration_schema.go && make && make readme && date +%H:%M:%S"

examples:
	cd examples/1; ortfodb in build out/database.json --config conf/ortfodb.yaml --scattered --write-progress out/.progress.json
