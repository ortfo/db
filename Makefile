.PHONY: configuration_schema.go

build:
	cd cmd/ortfodb; \
		go mod tidy; \
		go build;

install:
	sudo cp cmd/ortfodb/ortfodb ~/.local/bin/ortfodb
	sudo chmod +x ~/.local/bin/ortfodb

readme:
	./tools/build_readme.py

configuration_schema.go:
	./tools/build_configuration_schema_go.py

dev:
	filewatcher -I "{**.go,configuration.schema.json,_README.md}" -x configuration_schema.go "make configuration_schema.go && make && make readme && date +%H:%M:%S"
