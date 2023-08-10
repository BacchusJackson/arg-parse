#!/bin/bash

go run ../main.go help

BUILD_ARGS="version='v1.1.0' key1='some value for key1' key2='some other value \"key2\"'"

echo "Build Arguments -> $BUILD_ARGS"

export BUILD_CMD=$(arg-parse "docker build @@ -t myapp -f Dockerfile ." $BUILD_ARGS)

echo "Build Command -> $BUILD_CMD"

set -x
eval $BUILD_CMD

docker run myapp
