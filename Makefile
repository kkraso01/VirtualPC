.PHONY: build dev-up dev-reset test smoke

build:
	./scripts/build-binaries.sh

dev-up: build
	./scripts/dev-up.sh

dev-reset:
	./scripts/dev-reset.sh

test:
	go test ./...

smoke: build
	./scripts/smoke.sh
