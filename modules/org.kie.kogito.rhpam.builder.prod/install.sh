#!/bin/sh
set -e

cd $REMOTE_SOURCE_DIR/app
source $CACHITO_ENV_FILE && go build -a -o rhpam-kogito-operator-manager main.go
mkdir /workspace && cp $REMOTE_SOURCE_DIR/app/rhpam-kogito-operator-manager /workspace
