APP_FILE=./cmd/manager/main.go
IMAGE_REGISTRY=quay.io
APPLICATION_NAMESPACE=sbuvaneshkumar
REGISTRY_ORG=sbuvaneshkumar
REGISTRY_REPO=kogito-cloud-operator
IMAGE_LATEST_TAG=$(IMAGE_REGISTRY)/$(REGISTRY_ORG)/$(REGISTRY_REPO):latest
IMAGE_MASTER_TAG=$(IMAGE_REGISTRY)/$(REGISTRY_ORG)/$(REGISTRY_REPO):0.6.0
IMAGE_RELEASE_TAG=$(IMAGE_REGISTRY)/$(REGISTRY_ORG)/$(REGISTRY_REPO):$(CIRCLE_TAG)
# set CIRCLE_TAG (git release tag) in circleci env

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
build:
	./hack/go-build.sh

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

.PHONY: run-e2e
namespace = ""
tag = ""
native = "false"
maven_mirror = ""
image = ""
run-e2e:
	./hack/run-e2e.sh $(namespace) $(tag) $(native) $(maven_mirror) $(image)

.PHONY: run-e2e-cli
namespace = ""
tag = ""
native = "false"
maven_mirror = ""
skip_build = "false"
run-e2e-cli:
	./hack/run-e2e-cli.sh $(namespace) $(tag) $(native) $(maven_mirror) $(skip_build)

.PHONY: prepare-olm
version = ""
prepare-olm:
	./hack/pr-operatorhub.sh $(version)

.PHONY: code/build/linux
code/build/linux:
	env GOOS=linux GOARCH=amd64 go build $(APP_FILE)

.PHONY: image/build/master
image/build/master:
	@echo Building operator with the tag: $(IMAGE_MASTER_TAG)
	@docker login --username $(REGISTRY_USER) --password $(REGISTRY_PASS) https://registry.redhat.io
	operator-sdk build $(IMAGE_MASTER_TAG)

.PHONY: image/build/release
image/build/release:
	@echo Building operator with the tag: $(IMAGE_RELEASE_TAG)
	@docker login --username $(REGISTRY_USER) --password $(REGISTRY_PASS) https://registry.redhat.io
	operator-sdk build $(IMAGE_RELEASE_TAG)
	operator-sdk build $(IMAGE_LATEST_TAG)

.PHONY: image/push/master
image/push/master:
	@echo Pushing operator with tag $(IMAGE_MASTER_TAG) to $(IMAGE_REGISTRY)
	@docker login --username $(QUAY_USER) --password $(QUAY_PASS) quay.io
	docker push $(IMAGE_MASTER_TAG)

.PHONY: image/push/release
image/push/release:
	@echo Pushing operator with tag $(IMAGE_RELEASE_TAG) to $(IMAGE_REGISTRY)
	@docker login --username $(QUAY_USER) --password $(QUAY_PASS) quay.io
	docker push $(IMAGE_RELEASE_TAG)
	@echo Pushing operator with tag $(IMAGE_LATEST_TAG) to $(IMAGE_REGISTRY)
	docker push $(IMAGE_LATEST_TAG)
