# Current Operator version
VERSION ?= 1.8.1-snapshot
# Default bundle image tag
BUNDLE_IMG ?= quay.io/kiegroup/kogito-operator-bundle:$(VERSION)
# Default catalog image tag
CATALOG_IMG ?= quay.io/kiegroup/kogito-operator-catalog:$(VERSION)
# Default bundle image tag
PROFILING_IMG ?= quay.io/kiegroup/kogito-operator-profiling:$(VERSION)
# Options for 'bundle-build'
CHANNELS=alpha,1.x
BUNDLE_CHANNELS := --channels=$(CHANNELS)
DEFAULT_CHANNEL=1.x
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/kiegroup/kogito-operator:$(VERSION)
# Produce CRDs with v1 extension which is required by kubernetes v1.22+, The CRDs will stop working in kubernets <= v1.15
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Image tag to build the image with
IMAGE ?= $(IMG)

# Container runtime engine used for building the images
BUILDER ?= podman

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: generate manifests docker-build

profiling: generate manifests profiling-build

# Run tests
ENVTEST_ASSETS_DIR = $(shell pwd)/testbin
test: fmt lint
	./hack/go-test.sh

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go mod tidy
	./hack/addheaders.sh
	./hack/go-fmt.sh

lint:
	./hack/go-lint.sh

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	./hack/openapi.sh

# Build the docker image
docker-build:
	cekit -v build $(BUILDER)
	$(BUILDER) tag operator-runtime ${IMAGE}
# Push the docker image
docker-push:
	$(BUILDER) push ${IMAGE}

# Build the profiling docker image
profiling-build:
	cp image.yaml image.yaml.save && \
	cp profiling/image.yaml image.yaml && \
	cekit -v build $(BUILDER); \
	mv image.yaml.save image.yaml
	$(BUILDER) tag operator-runtime ${PROFILING_IMG}
# Push the profiling docker image
profiling-push:
	$(BUILDER) push ${PROFILING_IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v4@v4.0.5 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	$(BUILDER) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# Push the bundle image.
.PHONY: bundle-push
bundle-push:
	$(BUILDER) push ${BUNDLE_IMG}

# Build the catalog image.
.PHONY: catalog-build
catalog-build:
	opm index add -c ${BUILDER} --bundles ${BUNDLE_IMG}  --tag ${CATALOG_IMG}

# Push the catalog image.
.PHONY: catalog-push
catalog-push:
	$(BUILDER) push ${CATALOG_IMG}

generate-installer: generate manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/default > kogito-operator.yaml

generate-profiling-installer: generate manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(PROFILING_IMG)
	$(KUSTOMIZE) build config/profiling > profiling/kogito-operator-profiling.yaml
	$(KUSTOMIZE) build config/profiling-data-access > profiling/kogito-operator-profiling-data-access.yaml
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)

# Generate CSV
csv:
	operator-sdk generate kustomize manifests

vet: generate-installer generate-profiling-installer bundle
	go vet ./...


.PHONY: build-cli
release=false
version=""
build-cli:
	./hack/go-build-cli.sh $(release) $(version)

.PHONY: install-cli
install-cli:
	./hack/go-install-cli.sh

# Update bundle manifest files for test purposes, will override default image tag and remove the replaces field
.PHONY: update-bundle
update-bundle:
	./hack/update-bundle.sh ${IMAGE}

.PHONY: bump-version
new_version = ""
bump-version:
	./hack/bump-version.sh $(new_version)


.PHONY: deploy-operator-on-ocp
image ?= $2
deploy-operator-on-ocp:
	./hack/deploy-operator-on-ocp.sh $(image)

olm-tests:
	./hack/ci/run-olm-tests.sh

# Run this before any PR to make sure everything is updated, so CI won't fail
before-pr: vet test

#Run this to create a bundle dir structure in which OLM accepts. The bundle will be available in `build/_output/olm/<current-version>`
olm-manifests: bundle
	./hack/create-olm-manifests.sh

######
# Test proxy commands

TEST_DIR=test

.PHONY: run-tests
run-tests:
	@(cd $(TEST_DIR) && $(MAKE) $@)

.PHONY: build-examples-images
build-examples-images:
	@(cd $(TEST_DIR) && $(MAKE) $@)

.PHONY: run-smoke-tests
run-smoke-tests:
	@(cd $(TEST_DIR) && $(MAKE) $@)

.PHONY: build-smoke-examples-images
build-smoke-examples-images:
	@(cd $(TEST_DIR) && $(MAKE) $@)

.PHONY: run-performance-tests
run-performance-tests:
	@(cd $(TEST_DIR) && $(MAKE) $@)

.PHONY: build-performance-examples-images
build-performance-examples-images:
	@(cd $(TEST_DIR) && $(MAKE) $@)