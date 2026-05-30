.PHONY: build run release docker-build docker-up clean lint

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

clean:
	swift package clean

lint:
	swiftlint lint --strict
