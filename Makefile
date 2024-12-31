build:
	go build --race -o bin/$(shell basename $(PWD)) main.go

run: build
	bin/$(shell basename $(PWD))