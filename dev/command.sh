#!/bin/sh
# Timeout argument to be used in GitHub Workflows

if [ "$APP_CONTEXT" = "dev" ]; then CompileDaemon -log-prefix=false -build="go build" -command="./openslides-manage-service"; fi
if [ "$APP_CONTEXT" = "tests" ]; then sleep inf; fi
