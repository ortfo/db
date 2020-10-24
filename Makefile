.PHONY: configuration_schema.go

build:
	go build

install:
	sudo cp portfoliodb /usr/bin/portfoliodb
	sudo chmod +x /usr/bin/portfoliodb

readme:
	./tools/build_readme.py

configuration_schema.go:
	./tools/build_configuration_schema_go.py

dev:
	filewatcher -I "{**.go,configuration.schema.json,_README.md}" -x configuration_schema.go "make configuration_schema.go && make && make readme && date +%H:%M:%S"
