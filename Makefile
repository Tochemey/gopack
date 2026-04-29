SHELL := /bin/bash
.DEFAULT_GOAL := help

GO_VERSION ?= 1.25.7
GOLANGCI_LINT_VERSION ?= v2.10.1
BUF_VERSION ?= v1.65.0

IMAGE_NAME := gopack-dev
IMAGE_TAG  ?= $(GO_VERSION)
IMAGE      := $(IMAGE_NAME):$(IMAGE_TAG)

DOCKER_RUN_BASE := docker run --rm \
	-v "$(CURDIR):/app" \
	-w /app

DOCKER_RUN := $(DOCKER_RUN_BASE) $(IMAGE)

# Targets that need to launch sibling Docker containers (testcontainers).
# We mount the host docker socket and tell testcontainers-go to expose mapped
# ports on host.docker.internal so the test process can reach them.
DOCKER_RUN_DIND := $(DOCKER_RUN_BASE) \
	-v /var/run/docker.sock:/var/run/docker.sock \
	--add-host=host.docker.internal:host-gateway \
	-e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal \
	-e TESTCONTAINERS_RYUK_DISABLED=true \
	$(IMAGE)

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: image
image: ## Build (or rebuild) the dev/CI Docker image; Docker's layer cache makes this a no-op when nothing changed
	docker build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg BUF_VERSION=$(BUF_VERSION) \
		-t $(IMAGE) .

.PHONY: code
code: image ## Download Go module dependencies inside the build image
	$(DOCKER_RUN) go mod download -x

.PHONY: vendor
vendor: image ## Tidy and vendor Go modules
	$(DOCKER_RUN) sh -c "go mod tidy && go mod vendor"

.PHONY: lint
lint: vendor ## Run golangci-lint
	$(DOCKER_RUN) golangci-lint run --timeout 10m

.PHONY: local-test
local-test: vendor ## Run unit tests with race detector and coverage (uses Docker-in-Docker for testcontainers)
	$(DOCKER_RUN_DIND) go test -mod=vendor ./... -race -v -coverprofile=coverage.out -covermode=atomic -coverpkg=./...

.PHONY: test
test: local-test ## Alias for local-test (matches the Earthfile +test target)

.PHONY: protogen
protogen: image ## Generate protobuf code into test/data
	$(DOCKER_RUN) sh -c "buf generate --template buf.gen.yaml --path protos/test && rm -rf test/data && mv gen test/data"

.PHONY: pbs
pbs: protogen ## Alias for protogen (matches the Earthfile +pbs target)

.PHONY: clean
clean: ## Remove generated artifacts
	rm -f coverage.out
	rm -rf gen
