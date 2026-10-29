.PHONY: build run clean lint

IMAGE = smparkin/website:local

build:
	container build -t $(IMAGE) .

run:
	container run --rm -p 8080:8080 --rosetta $(IMAGE)

clean:
	swift package clean

lint:
	swiftlint lint --strict
