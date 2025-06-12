#!/bin/sh

if [ ! -z $dev   ]; CompileDaemon -log-prefix=false -build="go build ./cmd/server" -command="./server"; fi
if [ ! -z $tests ]; then make test; fi