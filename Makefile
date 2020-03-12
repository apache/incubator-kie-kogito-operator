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
tags=
concurrent=1
feature=
local=false
ci=
operator_image=
operator_tag=
cli_path=
deploy_uri=
services_image_version=
maven_mirror=
build_image_version=
build_s2i_image_tag=
build_runtime_image_tag=
examples_uri=
examples_ref=
timeout=240
debug=false
smoke=false
run-tests:
	./hack/run-tests.sh \
		--tags "${tags}" \
		--concurrent ${concurrent} \
		--feature ${feature} \
		--local ${local} \
		--ci ${ci} \
		--operator_image $(operator_image) \
		--operator_tag $(operator_tag) \
		--cli_path ${cli_path} \
		--deploy_uri ${deploy_uri} \
		--services_image_version ${services_image_version} \
		--maven_mirror $(maven_mirror) \
		--build_image_version ${build_image_version} \
		--build_s2i_image_tag ${build_s2i_image_tag} \
		--build_runtime_image_tag ${build_runtime_image_tag} \
		--examples_uri ${examples_uri} \
		--examples_ref ${examples_ref} \
		--timeout ${timeout} \
		--debug ${debug} \
		--smoke ${smoke}

.PHONY: run-smoke-tests
run-smoke-tests: 
	make run-tests smoke=true

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
