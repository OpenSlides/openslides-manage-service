all: openslides

build-dev:
	docker build . --target development --tag openslides-manage-dev

run-tests:
	docker build . --target testing --tag openslides-manage-test
	docker run openslides-manage-test

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
			chown build: /root/openslides ; \
		fi; \
		cp -p /root/openslides /build/"

.PHONY: openslides
