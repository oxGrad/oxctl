.PHONY: build test docker

build:
	go build -o bin/oxctl ./cmd/oxctl

test:
	go test ./...

docker:
	docker build -t oxctl:dev .
