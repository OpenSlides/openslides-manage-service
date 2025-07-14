#!/bin/bash

set -e

# Executes all tests. Should errors occur, CATCH will be set to 1, causing an erroneous exit code.

echo "########################################################################"
echo "###################### Run Tests and Linters ###########################"
echo "########################################################################"

# Parameters
while getopts "s" FLAG; do
    case "${FLAG}" in
    s) SKIP_BUILD=true ;;
    *) echo "Can't parse flag ${FLAG}" && break ;;
    esac
done

# Setup
CONTAINER_NAME="manage-tests"
IMAGE_TAG=openslides-manage-tests
LOCAL_PWD=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Safe Exit
trap 'docker stop "$CONTAINER_NAME" && docker rm "$CONTAINER_NAME"' EXIT

# Execution
if [ -z "$SKIP_BUILD" ]; then make build-tests &> /dev/null; fi
docker run -d --privileged --name "$CONTAINER_NAME" ${IMAGE_TAG}
docker exec "$CONTAINER_NAME" ./dev/container-tests.sh

# Linters
bash "$LOCAL_PWD"/run-lint.sh -s