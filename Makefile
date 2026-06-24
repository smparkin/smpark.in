.PHONY: build run clean lint sync-themes

IMAGE = smparkin/website:local

build:
	docker build -t $(IMAGE) .

run:
	docker run --rm -p 8080:8080 $(IMAGE)

clean:
	swift package clean

sync-themes:
	scripts/sync-themes.sh

lint:
	swiftlint lint --strict
