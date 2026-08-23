.PHONY: build test benzhi-docker
build:
	go build -o bin/reefctl ./cmd/reefctl
test:
	go test ./... -count=1
benzhi-docker:
	sh build_benzhi_docker.sh