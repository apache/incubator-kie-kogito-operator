APP_FILE=./cmd/manager/main.go
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
	./hack/go-test.sh

.PHONY: lint
lint:
	./hack/go-lint.sh
	#./hack/yaml-lint.sh

.PHONY: build
build:
	./hack/go-build.sh

.PHONY: build-cli
build-cli:
	./hack/go-build-cli.sh

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
run-e2e:
	./hack/run-e2e.sh $(namespace) $(tag) $(native) $(maven_mirror)

.PHONY: code/build/linux
code/build/linux:
	env GOOS=linux GOARCH=amd64 go build $(APP_FILE)

.PHONY: image/build/master
image/build/master:
	@echo Building operator
	@docker login --username $(REGISTRY_USER) --password $(REGISTRY_PASS) https://registry.redhat.io
	operator-sdk build quay.io/sbuvaneshkumar/kogito-cloud-operator

.PHONY: image/push/master
image/push/master:
	@echo Pushing operator 
	@docker login --username $(DOCKER_USER) --password $(DOCKER_PASS) quay.io
	docker push quay.io/sbuvaneshkumar/kogito-cloud-operator
