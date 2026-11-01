.PHONY: build run test lint docker-build docker-up container-build container-run clean upgrade

build:
	go build -o server .

run:
	go run .

test:
	go test -race ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

docker-build:
	docker buildx build --platform linux/amd64 -t smparkin/website:local --load .

docker-up:
	docker compose up

container-build:
	container build --platform linux/amd64 -t smparkin/website:local .

container-run:
	container run --rm -p 8080:8080 smparkin/website:local

clean:
	rm -f server

upgrade:
	go get -u ./... && go mod tidy
