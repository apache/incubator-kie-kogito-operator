# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

#export CGO_ENABLED:=0

.PHONY: all
all: build

.PHONY: mod
mod:
	./hack/go-mod.sh

.PHONY: format
format:
	./hack/go-fmt.sh

.PHONY: go-generate
go-generate: mod
	$(Q)go generate ./...

.PHONY: sdk-generate
sdk-generate: mod
	operator-sdk generate k8s

.PHONY: vet
vet:
	./hack/go-vet.sh

.PHONY: test
test:
	./hack/go-test.sh $(coverage)

.PHONY: lint
lint:
	./hack/go-lint.sh
	#./hack/yaml-lint.sh

.PHONY: build
image_registry=
image_name=
image_tag=
image_builder=
build:
	./hack/go-build.sh --image_registry ${image_registry} --image_name ${image_name} --image_tag ${image_tag} --image_builder ${image_builder}

.PHONY: deploy-operator-on-ocp
image=
deploy-operator-on-ocp:
	./hack/deploy-operator-on-ocp.sh $(image)

.PHONY: build-cli
release = false
version = ""
build-cli:
	./hack/go-build-cli.sh $(release) $(version)

.PHONY: install-cli
install-cli:
	./hack/go-install-cli.sh

.PHONY: clean
clean:
	rm -rf build/_output

.PHONY: addheaders
addheaders:
	./hack/addheaders.sh

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
# operator information
operator_image=
operator_tag=
# files/binaries
deploy_uri=
cli_path=
# runtime
services_image_version=
services_image_namespace=
services_image_registry=
data_index_image_tag=
jobs_service_image_tag=
management_console_image_tag=
# build
maven_mirror=
build_image_version=
build_image_namespace=
build_image_registry=
build_s2i_image_tag=
build_runtime_image_tag=
# examples repository
examples_uri=
examples_ref=
# dev options
show_scenarios=false
show_steps=false
dry_run=false
keep_namespace=false
disabled_crds_update=false
namespace_name=
run-tests:
	declare -a opts \
	&& if [ "${debug}" = "true" ]; then opts+=("--debug"); fi \
	&& if [ "${smoke}" = "true" ]; then opts+=("--smoke"); fi \
	&& if [ "${performance}" = "true" ]; then opts+=("--performance"); fi \
	&& if [ "${local}" = "true" ]; then opts+=("--local"); fi \
	&& if [ "${cr_deployment_only}" = "true" ]; then opts+=("--cr_deployment_only"); fi \
	&& if [ "${show_scenarios}" = "true" ]; then opts+=("--show_scenarios"); fi \
	&& if [ "${show_steps}" = "true" ]; then opts+=("--show_steps"); fi \
	&& if [ "${dry_run}" = "true" ]; then opts+=("--dry_run"); fi \
	&& if [ "${keep_namespace}" = "true" ]; then opts+=("--keep_namespace"); fi \
	&& if [ "${disabled_crds_update}" = "true" ]; then opts+=("--disabled_crds_update"); fi \
	&& if [ "${load_default_config}" = "true" ]; then opts+=("--load_default_config"); fi \
	&& opts_str=$$(IFS=' ' ; echo "$${opts[*]}") \
	&& ./hack/run-tests.sh \
		--feature ${feature} \
		--tags "${tags}" \
		--concurrent ${concurrent} \
		--timeout ${timeout} \
		--ci ${ci} \
		--operator_image $(operator_image) \
		--operator_tag $(operator_tag) \
		--deploy_uri ${deploy_uri} \
		--cli_path ${cli_path} \
		--services_image_version ${services_image_version} \
		--services_image_namespace ${services_image_namespace} \
		--services_image_registry ${services_image_registry} \
		--data_index_image_tag ${data_index_image_tag} \
		--jobs_service_image_tag ${jobs_service_image_tag} \
		--management_console_image_tag ${management_console_image_tag} \
		--maven_mirror $(maven_mirror) \
		--build_image_version ${build_image_version} \
		--build_image_namespace ${build_image_namespace} \
		--build_image_registry ${build_image_registry} \
		--build_s2i_image_tag ${build_s2i_image_tag} \
		--build_runtime_image_tag ${build_runtime_image_tag} \
		--examples_uri ${examples_uri} \
		--examples_ref ${examples_ref} \
		--namespace_name ${namespace_name} \
		--load_factor ${load_factor} \
		$${opts_str}

.PHONY: run-smoke-tests
run-smoke-tests: 
	make run-tests smoke=true

.PHONY: run-performance-tests
run-performance-tests:
	make run-tests performance=true

.PHONY: prepare-olm
version = ""
prepare-olm:
	./hack/pr-operatorhub.sh $(version)

.PHONY: bump-version
old_version = ""
new_version = ""
bump-version:
	./hack/bump-version.sh $(old_version) $(new_version)

.PHONY: scorecard
scorecard:
	./hack/scorecard.sh
