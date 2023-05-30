# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 1.39.1-snapshot

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
CHANNELS ?= alpha,1.x
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL ?= 1.x
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# quay.io/kiegroup/kogito-operator-bundle:$VERSION and quay.io/kiegroup/kogito-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/kiegroup/kogito-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
    BUNDLE_GEN_FLAGS += --use-image-digests
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# PROFILING_IMG defines the image:tag used for the profiling.
# It is used to catch the coverage with BDD test.
PROFILING_IMG ?= $(IMAGE_TAG_BASE)-profiling:$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Container runtime engine used for building the images
BUILDER ?= podman

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

print-%  : ; @echo $($*)

all: generate manifests container-build
	echo "calling APP all ##################################"

profiling: generate manifests profiling-build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

# Run tests
ENVTEST_ASSETS_DIR = $(shell pwd)/testbin
test: fmt lint
	./hack/go-test.sh

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	echo "calling APP manifests ##################################"
	./hack/kogito-module-api.sh --disable
	$(CONTROLLER_GEN) crd paths="./apis/app/..." output:crd:artifacts:config=config/crd/app/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./controllers/app" output:rbac:artifacts:config=config/rbac/app
	$(CONTROLLER_GEN) webhook paths="./..."
	./hack/kogito-module-api.sh --enable

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	echo "calling APP generate ##################################"
	./hack/kogito-module-api.sh --disable
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths=./...
	./hack/kogito-module-api.sh --enable
	./hack/client-gen.sh

fmt: ## Run go fmt against code.
	go mod tidy
	./hack/addheaders.sh
	./hack/go-fmt.sh

lint:
	./hack/go-lint.sh

# Run vet
.PHONY: vet
vet: generate-installer generate-profiling-installer bundle ## Run go vet against code.
	go vet ./...

# Generate CSV
csv: install-operator-sdk
	./hack/kogito-module-api.sh --disable
	operator-sdk generate kustomize manifests  --apis-dir=apis/app --input-dir=./config/manifests/app --output-dir=./config/manifests/app --package=kogito-operator -q
	./hack/kogito-module-api.sh --enable

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

container-build: ## Build the docker image
	echo "calling APP container-build ##################################"
	cekit -v --descriptor kogito-image.yaml build $(BUILDER)
	$(BUILDER) tag operator-runtime ${IMG}

container-push: ## Push the docker image
	$(BUILDER) push ${IMG}

profiling-build: ## Build the profiling docker image
	cp kogito-image.yaml kogito-image.yaml.save && \
	cp profiling/image.yaml kogito-image.yaml && \
	cekit -v --descriptor kogito-image.yaml build $(BUILDER); \
	mv kogito-image.yaml.save kogito-image.yaml
	$(BUILDER) tag operator-runtime ${PROFILING_IMG}

profiling-push: ## Push the profiling docker image
	$(BUILDER) push ${PROFILING_IMG}

.PHONY: build-cli
release=false
version=""
build-cli:
	./hack/go-build-cli.sh $(release) $(version)

.PHONY: install-cli
install-cli:
	./hack/go-install-cli.sh

ifndef ignore-not-found
  ignore-not-found = false
endif
##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.2
CONTROLLER_TOOLS_VERSION ?= v0.8.0

# KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	rm -rfv $(LOCALBIN)/kustomize
	GO111MODULE=on GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v4@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

bundle: manifests kustomize install-operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	echo "calling APP bundle ##################################"
	./hack/kogito-module-api.sh --disable
	operator-sdk generate kustomize manifests  --apis-dir=apis/app --input-dir=./config/manifests/app --output-dir=./config/manifests/app --package=kogito-operator -q
	cd config/manager/app && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests/app | operator-sdk generate bundle --package=kogito-operator --output-dir=bundle/app $(BUNDLE_GEN_FLAGS)
	operator-sdk bundle validate ./bundle/app
	./hack/kogito-module-api.sh --enable

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(BUILDER) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) container-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool ${BUILDER} --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) container-push IMG=$(CATALOG_IMG)


generate-installer: generate manifests kustomize
	echo "calling APP generate-installer ##################################"
	cd config/manager/app && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/default/app > kogito-operator.yaml

generate-profiling-installer: generate manifests kustomize
	echo "calling APP generate-profiling-installer ##################################"
	cd config/manager/app && $(KUSTOMIZE) edit set image controller=$(PROFILING_IMG)
	$(KUSTOMIZE) build config/profiling > profiling/kogito-operator-profiling.yaml
	$(KUSTOMIZE) build config/profiling-data-access > profiling/kogito-operator-profiling-data-access.yaml
	cd config/manager/app && $(KUSTOMIZE) edit set image controller=$(IMG)

# Update bundle manifest files for test purposes, will override default image tag and remove the replaces field
.PHONY: update-bundle
update-bundle:
	./hack/update-bundle.sh ${IMG}

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

install-operator-sdk:
	./hack/ci/install-operator-sdk.sh

uninstall-operator-sdk:
	./hack/ci/uninstall-operator-sdk.sh

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