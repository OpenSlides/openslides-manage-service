build-dev:
	docker build . --target development --tag openslides-manage-dev

run-tests:
	docker build . --target testing --tag openslides-manage-test
	docker run openslides-manage-test

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/manage.proto
