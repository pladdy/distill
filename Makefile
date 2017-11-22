.PHONY: build

all: build

build:
	go build

clean:
	rm -f distill

test:
	go test -cover -v ./...
