.PHONY: build dev-up dev-reset test smoke e2e soak

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

e2e: build
	VPC_E2E_REAL_HOST=1 go test ./tests/e2e -run TestRealHostE2E -v

soak: build
	VPC_SOAK_REAL_HOST=1 go test ./tests/reliability -run TestSoak -v
