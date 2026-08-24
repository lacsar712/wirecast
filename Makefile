.PHONY: build test benzhi-docker
build:
	go build -o bin/graindry ./cmd/graindry
test:
	go test ./... -count=1
benzhi-docker:
	sh build_benzhi_docker.sh
