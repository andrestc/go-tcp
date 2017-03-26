build:
	go build -o ./bin/go-tcp

run: build
	./bin/go-tcp