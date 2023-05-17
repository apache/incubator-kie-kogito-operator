FROM scratch

# Core bundle labels.
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=rhpam-kogito-operator
LABEL operators.operatorframework.io.bundle.channels.v1=7.x
LABEL operators.operatorframework.io.bundle.channel.default.v1=7.x
LABEL operators.operatorframework.io.metrics.builder=operator-sdk-v1.21.0
LABEL operators.operatorframework.io.metrics.mediatype.v1=metrics+v1
LABEL operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v3

# Labels for testing.
LABEL operators.operatorframework.io.test.mediatype.v1=scorecard+v1
LABEL operators.operatorframework.io.test.config.v1=tests/scorecard/

# Copy files to locations specified by labels.
COPY bundle/rhpam/manifests /manifests/
COPY bundle/rhpam/metadata /metadata/
COPY bundle/rhpam/tests/scorecard /tests/scorecard/
