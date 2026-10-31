.PHONY: build run release docker-build docker-up container-build container-run clean lint

build:
	swift build

run:
	swift run

release:
	swift build -c release

docker-build:
	docker buildx build --platform linux/amd64 -t smparkin/website:local --load .

docker-up:
	docker compose up

# Apple's container CLI; the builder VM needs more than its 2 CPU / 2 GB default for a release build
container-build:
	container build --platform linux/amd64 --cpus 8 --memory 8G -t smparkin/website:local .

container-run:
	container run --rm -p 8080:8080 smparkin/website:local

clean:
	swift package clean

lint:
	swiftlint lint --strict
