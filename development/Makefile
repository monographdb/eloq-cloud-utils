# EloqCloud Development Base Image Makefile

# Image configuration
IMAGE = eloqdata/eloqcloud-dev:go-1.26

# Build platforms
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: ## Build the base image for local platform
	docker build -t $(IMAGE) .

.PHONY: build-multiarch
build-multiarch: ## Build multi-architecture image (requires buildx)
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(IMAGE) \
		.

##@ Push

.PHONY: push
push: ## Push the image to registry
	docker push $(IMAGE)

.PHONY: build-push
build-push: ## Build and push multi-architecture image
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-t $(IMAGE) \
		.

##@ Test

.PHONY: test
test: build ## Test the built image
	@echo "Testing image: $(IMAGE)"
	docker run --rm $(IMAGE) go version
	docker run --rm $(IMAGE) protoc --version
	docker run --rm $(IMAGE) protoc-gen-go --version
	docker run --rm $(IMAGE) protoc-gen-go-grpc --version
	docker run --rm $(IMAGE) swag --version

##@ Utility

.PHONY: clean
clean: ## Remove local images
	docker rmi $(IMAGE) || true

.PHONY: shell
shell: ## Run interactive shell in the image
	docker run --rm -it -v $(PWD):/workspace $(IMAGE) /bin/bash
