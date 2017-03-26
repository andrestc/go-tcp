build:
	go build -o ./bin/go-tcp

run: build
	sudo ./bin/go-tcp