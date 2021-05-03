all: test

build-dev:
	docker build . --target development --tag openslides-manage-dev

run-tests:
	docker build . --target testing --tag openslides-manage-test
	docker run openslides-manage-test

protoc:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative \
	proto/manage.proto

go-build:
	go build ./cmd/server
	go build ./cmd/manage

test:
	# Attention: This steps should be the same as in .github/workflows/test.yml.
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go install golang.org/x/lint/golint@latest
	golint -set_exit_status ./...
	go test -timeout 10s -race ./...
