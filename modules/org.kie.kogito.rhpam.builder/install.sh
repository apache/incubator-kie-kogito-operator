#!/bin/sh
set -e

if [ -n "$CACHITO_ENV_FILE" ]; then
  source $CACHITO_ENV_FILE && go build -a -o rhpam-kogito-operator-manager main.go
  cp $REMOTE_SOURCE_DIR/app/rhpam-kogito-operator-manager /workspace
else
  cd /workspace
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o rhpam-kogito-operator-manager main.go;
fi
