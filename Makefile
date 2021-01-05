# Current Operator version
VERSION ?= 2.0.0-snapshot
# Default bundle image tag
BUNDLE_IMG ?= quay.io/kiegroup/kogito-cloud-operator-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/kiegroup/kogito-cloud-operator:$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

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
ifeq ($(BUILDER),buildah)
	$(BUILDER) -t ${IMAGE} bud .
else
	$(BUILDER) build . -t ${IMAGE}
endif

# Push the docker image
docker-push:
	$(BUILDER) push ${IMAGE}


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
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
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

generate-installer: generate manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/default > kogito-operator.yaml

# Generate CSV
csv:
	operator-sdk generate kustomize manifests

vet: generate-installer bundle
	go vet ./...


.PHONY: build-cli
release=false
version=""
build-cli:
	./hack/go-build-cli.sh $(release) $(version)

.PHONY: install-cli
install-cli:
	./hack/go-install-cli.sh

.PHONY: run-tests
# tests configuration
feature=
tags=
concurrent=1
timeout=240
debug=false
smoke=false
performance=false
load_factor=1
local=false
ci=
cr_deployment_only=false
load_default_config=false
container_engine=
domain_suffix=
image_cache_mode=
http_retry_nb=
olm_namespace=
# operator information
operator_image=
operator_tag=
operator_namespaced=false
# files/binaries
operator_yaml_uri=
cli_path=
# runtime
services_image_registry=
services_image_namespace=
services_image_name_suffix=
services_image_version=
data_index_image_tag=
explainability_image_tag=
jobs_service_image_tag=
management_console_image_tag=
task_console_image_tag=
trusty_image_tag=
trusty_ui_image_tag=
runtime_application_image_registry=
runtime_application_image_namespace=
runtime_application_image_name_suffix=
runtime_application_image_version=
# build
custom_maven_repo=
custom_maven_repo_replace_default=false
maven_mirror=
maven_ignore_self_signed_certificate=false
build_image_registry=
build_image_namespace=
build_image_name_suffix=
build_image_version=
build_s2i_image_tag=
build_runtime_image_tag=
disable_maven_native_build_container=false
# examples repository
examples_uri=
examples_ref=
# dev options
show_scenarios=false
show_steps=false
dry_run=false
keep_namespace=false
namespace_name=
local_cluster=false
run-tests:
	declare -a opts \
	&& if [ "${debug}" = "true" ]; then opts+=("--debug"); fi \
	&& if [ "${smoke}" = "true" ]; then opts+=("--smoke"); fi \
	&& if [ "${performance}" = "true" ]; then opts+=("--performance"); fi \
	&& if [ "${local}" = "true" ]; then opts+=("--local"); fi \
	&& if [ "${local_cluster}" = "true" ]; then opts+=("--local_cluster"); fi \
	&& if [ "${cr_deployment_only}" = "true" ]; then opts+=("--cr_deployment_only"); fi \
	&& if [ "${show_scenarios}" = "true" ]; then opts+=("--show_scenarios"); fi \
	&& if [ "${show_steps}" = "true" ]; then opts+=("--show_steps"); fi \
	&& if [ "${dry_run}" = "true" ]; then opts+=("--dry_run"); fi \
	&& if [ "${keep_namespace}" = "true" ]; then opts+=("--keep_namespace"); fi \
	&& if [ "${load_default_config}" = "true" ]; then opts+=("--load_default_config"); fi \
	&& if [ "${maven_ignore_self_signed_certificate}" = "true" ]; then opts+=("--maven_ignore_self_signed_certificate"); fi \
	&& if [ "${disable_maven_native_build_container}" = "true" ]; then opts+=("--disable_maven_native_build_container"); fi \
	&& if [ "${custom_maven_repo_replace_default}" = "true" ]; then opts+=("--custom_maven_repo_replace_default"); fi \
	&& if [ "${operator_namespaced}" = "true" ]; then opts+=("--operator_namespaced"); fi \
	&& opts_str=$$(IFS=' ' ; echo "$${opts[*]}") \
	&& ./hack/run-tests.sh \
		--feature ${feature} \
		--tags "${tags}" \
		--concurrent ${concurrent} \
		--timeout ${timeout} \
		--ci ${ci} \
		--operator_image $(operator_image) \
		--operator_tag $(operator_tag) \
		--operator_yaml_uri ${operator_yaml_uri} \
		--cli_path ${cli_path} \
		--services_image_registry ${services_image_registry} \
		--services_image_namespace ${services_image_namespace} \
		--services_image_name_suffix ${services_image_name_suffix} \
		--services_image_version ${services_image_version} \
		--data_index_image_tag ${data_index_image_tag} \
		--explainability_image_tag ${explainability_image_tag} \
		--jobs_service_image_tag ${jobs_service_image_tag} \
		--management_console_image_tag ${management_console_image_tag} \
		--task_console_image_tag ${task_console_image_tag} \
		--trusty_image_tag ${trusty_image_tag} \
		--trusty_ui_image_tag ${trusty_ui_image_tag} \
		--runtime_application_image_registry ${runtime_application_image_registry} \
		--runtime_application_image_namespace ${runtime_application_image_namespace} \
		--runtime_application_image_name_suffix ${runtime_application_image_name_suffix} \
		--runtime_application_image_version ${runtime_application_image_version} \
		--custom_maven_repo $(custom_maven_repo) \
		--maven_mirror $(maven_mirror) \
		--build_image_registry ${build_image_registry} \
		--build_image_namespace ${build_image_namespace} \
		--build_image_name_suffix ${build_image_name_suffix} \
		--build_image_version ${build_image_version} \
		--build_s2i_image_tag ${build_s2i_image_tag} \
		--build_runtime_image_tag ${build_runtime_image_tag} \
		--examples_uri ${examples_uri} \
		--examples_ref ${examples_ref} \
		--namespace_name ${namespace_name} \
		--load_factor ${load_factor} \
		--container_engine ${container_engine} \
		--domain_suffix ${domain_suffix} \
		--image_cache_mode ${image_cache_mode} \
		--http_retry_nb ${http_retry_nb} \
		--olm_namespace ${olm_namespace} \
		$${opts_str}

.PHONY: run-smoke-tests
run-smoke-tests:
	make run-tests smoke=true

.PHONY: run-performance-tests
run-performance-tests:
	make run-tests performance=true

.PHONY: build-examples-images
build-examples-images:
	make run-tests feature=scripts/examples cr_deployment_only=true

.PHONY: bump-version
new_version = ""
bump-version:
	./hack/bump-version.sh $(new_version)


.PHONY: deploy-operator-on-ocp
image ?= $2
deploy-operator-on-ocp:
	./hack/deploy-operator-on-ocp.sh $(image)