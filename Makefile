all: openslides

build-aio:
	@if [ -z "${submodule}" ] ; then \
		echo "Please provide the name of the submodule service to build (submodule=<submodule service name>)"; \
		exit 1; \
	fi

	@if [ "${context}" != "prod" -a "${context}" != "dev" -a "${context}" != "tests" ] ; then \
		echo "Please provide a context for this build (context=<desired_context> , possible options: prod, dev, tests)"; \
		exit 1; \
	fi

	echo "Building submodule '${submodule}' for ${context} context"

	@docker build -f ./Dockerfile.AIO ./ --tag openslides-${submodule}-${context} --build-arg CONTEXT=${context} --target ${context} ${args}

build-dev:
	make build-aio context=dev submodule=manage

#	docker build . --target development --tag openslides-manage-dev

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
			chown build: /root/openslides ; \
		fi; \
		cp -p /root/openslides /build/"

.PHONY: openslides
