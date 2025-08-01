SERVICE=manage

build-prod:
	docker build ./ --tag "openslides-$(SERVICE)" --build-arg CONTEXT="prod" --target "prod"

build-dev:
	docker build ./ --tag "openslides-$(SERVICE)-dev" --build-arg CONTEXT="dev" --target "dev"

build-test:
	docker build ./ --tag "openslides-$(SERVICE)-tests" --build-arg CONTEXT="tests" --target "tests"

all: openslides

run-tests:
	bash dev/run-tests.sh

test:
	# Attention: This steps should be the same as in .github/workflows/test.yml.
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go install golang.org/x/lint/golint@latest
	golint -set_exit_status ./...
	go test -timeout 10s -race ./...

go-build:
	go build ./cmd/openslides

protoc:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative \
	proto/manage.proto

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))

openslides:
	docker build . --target builder --tag openslides-manage-builder
	docker run --interactive --tty --volume $(dir $(mkfile_path)):/build/ --rm openslides-manage-builder sh -c " \
		if [ $(shell whoami) != root ]; then \
			addgroup -g $(shell id -g) build ; \
			adduser -u $(shell id -u) -G build -D build ; \
			chown build: /app/openslides ; \
		fi; \
		cp -p /app/openslides /build/"

.PHONY: openslides
