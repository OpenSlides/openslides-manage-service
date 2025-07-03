#!/bin/bash

# Executes all linters. Should errors occur, CATCH will be set to 1, causing an erroneous exit code.

echo "########################################################################"
echo "###################### Run Linters #####################################"
echo "########################################################################"

# Parameters
while getopts "lscp" FLAG; do
    case "${FLAG}" in
    l) LOCAL=true ;;
    s) SKIP_BUILD=true ;;
    c) SKIP_CONTAINER_UP=true ;;
    p) PERSIST_CONTAINERS=true ;;
    *) echo "Can't parse flag ${FLAG}" && break ;;
    esac
done

# Setup
IMAGE_TAG=openslides-manage-tests
CATCH=0
DOCKER_EXEC="docker exec manage-test"

# Optionally build image
if [ -z "$SKIP_BUILD" ]; then make build-tests || CATCH=1; fi

# Execution
if [ -z "$LOCAL" ]
then
    # Container Mode
    if [ -z "$SKIP_CONTAINER_UP" ]; then docker run -d -t --name manage-test ${IMAGE_TAG} || CATCH=1; fi
    eval "$DOCKER_EXEC go vet ./..." || CATCH=1
    eval "$DOCKER_EXEC golint -set_exit_status ./..." || CATCH=1
    eval "$DOCKER_EXEC gofmt -l ." || CATCH=1
else
    # Local Mode
    go vet ./... || CATCH=1
    golint -set_exit_status ./... || CATCH=1
    gofmt -l -s -w . || CATCH=1
fi

if [ -z "$PERSIST_CONTAINERS" ] && [ -z "$SKIP_CONTAINER_UP" ]; then docker stop manage-test && docker rm manage-test || CATCH=1; fi

exit $CATCH