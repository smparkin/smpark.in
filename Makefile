.PHONY: build run release docker-build docker-up clean

build:
	swift build

run:
	swift run

release:
	swift build -c release

docker-build:
	docker build -t sparkii/smparkin:latest .

docker-up:
	docker compose up

clean:
	swift package clean
