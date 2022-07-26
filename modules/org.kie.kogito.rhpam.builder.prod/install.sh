#!/bin/sh
set -e

cd $REMOTE_SOURCE_DIR/app
source $CACHITO_ENV_FILE && go build -a -o bamoe-kogito-operator-manager main.go
mkdir /workspace && cp $REMOTE_SOURCE_DIR/app/bamoe-kogito-operator-manager /workspace
