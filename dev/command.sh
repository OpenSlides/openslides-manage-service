#!/bin/sh
# Timeout argument to be used in GitHub Workflows

if [ "$APP_CONTEXT" = "dev" ]; then exec CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"; fi
if [ "$APP_CONTEXT" = "tests" ]; then sleep inf; fi
