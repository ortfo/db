build:
	go build

install:
	sudo cp portfoliodb /usr/bin/portfoliodb
	sudo chmod +x /usr/bin/portfoliodb

dev:
	filewatcher -I "**.go" "make && date +%H:%M:%S"
