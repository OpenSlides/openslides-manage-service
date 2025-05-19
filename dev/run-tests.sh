#!/bin/bash

# Executes all tests. Should errors occur, CATCH will be set to 1, causing an erronous exit code.

echo "########################################################################"
echo "###################### Start full system tests #########################"
echo "########################################################################"

IMAGE_TAG=openslides-manage-tests
CATCH=0
PERSIST_CONTAINERS=$1

docker build -f ./Dockerfile.AIO ./ --tag ${IMAGE_TAG} --build-arg CONTEXT=tests --target tests || CATCH=1
docker run ${IMAGE_TAG} || CATCH=1

if [ -z $PERSIST_CONTAINERS ]; then docker stop $(docker ps -a -q --filter ancestor=${IMAGE_TAG} --format="{{.ID}}") || CATCH=1; fi

exit $CATCH