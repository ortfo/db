build:
	go build

install:
	sudo cp portfoliodb /usr/bin/portfoliodb

dev:
	filewatcher -I "**.go" "make"
