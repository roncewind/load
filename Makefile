# Makefile that builds go-hello-world, a "go" program.

# "Simple expanded" variables (':=')

# PROGRAM_NAME is the name of the GIT repository.
PROGRAM_NAME := $(shell basename `git rev-parse --show-toplevel`)
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIRECTORY := $(dir $(MAKEFILE_PATH))
TARGET_DIRECTORY := $(MAKEFILE_DIRECTORY)/target
DOCKER_CONTAINER_NAME := $(PROGRAM_NAME)
DOCKER_IMAGE_NAME := dockter/$(PROGRAM_NAME)
DOCKER_BUILD_IMAGE_NAME := $(DOCKER_IMAGE_NAME)-build
BUILD_VERSION := $(shell git describe --always --tags --abbrev=0 --dirty)
BUILD_TAG := $(shell git describe --always --tags --abbrev=0)
BUILD_ITERATION := $(shell git log $(BUILD_TAG)..HEAD --oneline | wc -l | sed 's/^ *//')
GIT_REMOTE_URL := $(shell git config --get remote.origin.url)
GO_PACKAGE_NAME := $(shell echo $(GIT_REMOTE_URL) | sed -e 's|^git@github.com:|github.com/|' -e 's|\.git$$||')

# Recursive assignment ('=')

CC = gcc
# Conditional assignment. ('?=')

SENZING_G2_DIR ?= /opt/senzing/g2
SENZING_DATABASE_URL ?= postgresql://postgres:postgres@127.0.0.1:5432/G2

# The first "make" target runs as default.

.PHONY: default
default: help

# -----------------------------------------------------------------------------
# Export environment variables.
# -----------------------------------------------------------------------------

.EXPORT_ALL_VARIABLES:

# Flags for the C compiler

# Flags for

LD_LIBRARY_PATH = ${SENZING_G2_DIR}/lib

CGO_LDFLAGS = \
	-L${SENZING_G2_DIR}/lib \
	-lG2

# ---- Linux ------------------------------------------------------------------

target/linux:
	@mkdir -p $(TARGET_DIRECTORY)/linux || true


target/linux/$(PROGRAM_NAME): target/linux
	GOOS=linux \
	GOARCH=amd64 \
	go build \
		-a \
		-ldflags " \
			-X main.programName=${PROGRAM_NAME} \
			-X main.buildVersion=${BUILD_VERSION} \
			-X main.buildIteration=${BUILD_ITERATION} \
			" \
		-o $(TARGET_DIRECTORY)/linux/$(PROGRAM_NAME)

target/scratch:
	@mkdir -p $(TARGET_DIRECTORY)/scratch || true


target/linux/go-hello-senzing-dynamic: target/linux
	GOOS=linux \
	GOARCH=amd64 \
	go build \
		-a \
		-ldflags " \
			-X main.programName=${PROGRAM_NAME} \
			-X main.buildVersion=${BUILD_VERSION} \
			-X main.buildIteration=${BUILD_ITERATION} \
			" \
		-o $(TARGET_DIRECTORY)/linux/go-hello-senzing-dynamic


target/linux/go-hello-senzing-static: target/linux
	GOOS=linux \
	GOARCH=amd64 \
	go build \
		-a \
		-ldflags " \
			-X main.programName=${PROGRAM_NAME} \
			-X main.buildVersion=${BUILD_VERSION} \
			-X main.buildIteration=${BUILD_ITERATION} \
			-extldflags \"-static\" \
			" \
		-o $(TARGET_DIRECTORY)/linux/go-hello-senzing-static


# Can't use -extldflags \"-static\" because .a files are needed.
target/scratch/senzing: target/scratch dependencies
	GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build \
		-a \
		-installsuffix cgo \
		-ldflags " \
			-s \
			-w \
			-X main.programName=${PROGRAM_NAME} \
			-X main.buildVersion=${BUILD_VERSION} \
			-X main.buildIteration=${BUILD_ITERATION} \
			" \
		-o $(TARGET_DIRECTORY)/scratch/senzing


target/linux/go-hello-senzing-dynamicXX:
	GOOS=linux GOARCH=amd64 \
		go build \
			-a \
			-ldflags " \
				-X main.programName=${PROGRAM_NAME} \
				-X main.buildVersion=${BUILD_VERSION} \
				-X main.buildIteration=${BUILD_ITERATION} \
	    	" \
			${GO_PACKAGE_NAME}
	@mkdir -p $(TARGET_DIRECTORY)/linux || true
	@mv $(PROGRAM_NAME) $(TARGET_DIRECTORY)/linux/go-hello-senzing-dynamic

# -----------------------------------------------------------------------------
# Build
#   Notes:
#     "-a" needed to incorporate changes to C files.
# -----------------------------------------------------------------------------

.PHONY: dependencies
dependencies:
	@go get -u ./...
	@go get -t -u ./...
	@go mod tidy


.PHONY: build
build: dependencies \
	target/linux/$(PROGRAM_NAME)
#	target/linux/go-hello-senzing-dynamic
#	target/linux/go-hello-senzing-static \
#	target/scratch/senzing
#	target/linux/go-hello-senzing-static
#	build-macos \
#	build-scratch \
#	build-windows


.PHONY: build-macos
build-macos:
	@GOOS=darwin \
	GOARCH=amd64 \
	go build \
	  -ldflags \
	    "-X main.programName=${PROGRAM_NAME} \
	     -X main.buildVersion=${BUILD_VERSION} \
	     -X main.buildIteration=${BUILD_ITERATION} \
	     -X github.com/docktermj/go-hello-world-module.helloName=${HELLO_NAME} \
	    " \
	  -o $(GO_PACKAGE_NAME)
	@mkdir -p $(TARGET_DIRECTORY)/darwin || true
	@mv $(GO_PACKAGE_NAME) $(TARGET_DIRECTORY)/darwin


.PHONY: build-windows
build-windows:
	@GOOS=windows \
	GOARCH=amd64 \
	go build \
	  -ldflags \
	    "-X main.programName=${PROGRAM_NAME} \
	     -X main.buildVersion=${BUILD_VERSION} \
	     -X main.buildIteration=${BUILD_ITERATION} \
	     -X github.com/docktermj/go-hello-world-module.helloName=${HELLO_NAME} \
	    " \
	  -o $(GO_PACKAGE_NAME).exe
	@mkdir -p $(TARGET_DIRECTORY)/windows || true
	@mv $(GO_PACKAGE_NAME).exe $(TARGET_DIRECTORY)/windows

# -----------------------------------------------------------------------------
# Test
# -----------------------------------------------------------------------------

.PHONY: test
test:
#	@go test -v $(GO_PACKAGE_NAME)/...
#	@go test -v $(GO_PACKAGE_NAME)/g2diagnostic
#	@go test -v $(GO_PACKAGE_NAME)/g2engine
#	@go test -v $(GO_PACKAGE_NAME)/g2config
	@go test -v $(GO_PACKAGE_NAME)/g2configmgr

# -----------------------------------------------------------------------------
# Run
# -----------------------------------------------------------------------------

.PHONY: run
run:
#	@target/linux/$(PROGRAM_NAME)
	GOOS=linux \
	GOARCH=amd64 \
	go run . \
		--inputURL "amqp://guest:guest@192.168.6.96:5672"

# -----------------------------------------------------------------------------
# docker-build
#  - https://docs.docker.com/engine/reference/commandline/build/
# -----------------------------------------------------------------------------

.PHONY: docker-build
docker-build:
	@docker build \
		--build-arg BUILD_ITERATION=$(BUILD_ITERATION) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg GO_PACKAGE_NAME=$(GO_PACKAGE_NAME) \
		--build-arg PROGRAM_NAME=$(PROGRAM_NAME) \
		--file Dockerfile \
		--tag $(DOCKER_IMAGE_NAME) \
		--tag $(DOCKER_IMAGE_NAME):$(BUILD_VERSION) \
		.

.PHONY: docker-builder
docker-builder:
	@docker build \
		--build-arg BUILD_ITERATION=$(BUILD_ITERATION) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg GO_PACKAGE_NAME=$(GO_PACKAGE_NAME) \
		--build-arg PROGRAM_NAME=$(PROGRAM_NAME) \
		--file Dockerfile \
		--tag $(DOCKER_IMAGE_NAME) \
		--tag $(DOCKER_IMAGE_NAME):$(BUILD_VERSION) \
		--target go_builder\
		.


.PHONY: docker-build-package
docker-build-package:
	@docker build \
		--build-arg BUILD_ITERATION=$(BUILD_ITERATION) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg GO_PACKAGE_NAME=$(GO_PACKAGE_NAME) \
		--build-arg PROGRAM_NAME=$(PROGRAM_NAME) \
		--file package.Dockerfile \
		--no-cache \
		--tag $(DOCKER_BUILD_IMAGE_NAME) \
		.


.PHONY: docker-run
docker-run:
	@docker run \
	    --interactive \
	    --tty \
	    --name $(DOCKER_CONTAINER_NAME) \
	    $(DOCKER_IMAGE_NAME)

# -----------------------------------------------------------------------------
# Package
# -----------------------------------------------------------------------------

.PHONY: package
package: docker-build-package
	@mkdir -p $(TARGET_DIRECTORY) || true
	@CONTAINER_ID=$$(docker create $(DOCKER_BUILD_IMAGE_NAME)); \
	docker cp $$CONTAINER_ID:/output/. $(TARGET_DIRECTORY)/; \
	docker rm -v $$CONTAINER_ID

# -----------------------------------------------------------------------------
# Run
# -----------------------------------------------------------------------------

.PHONY: run-linux-dynamic
run-linux-dynamic:
	@target/linux/go-hello-senzing-dynamic

# -----------------------------------------------------------------------------
# Utility targets
# -----------------------------------------------------------------------------

.PHONY: clean
clean:
	@go clean -cache
	@docker rm --force $(DOCKER_CONTAINER_NAME) 2> /dev/null || true
	@docker rmi --force $(DOCKER_IMAGE_NAME) $(DOCKER_BUILD_IMAGE_NAME) 2> /dev/null || true
	@rm -rf $(TARGET_DIRECTORY) || true
	@rm -f $(GOPATH)/bin/$(PROGRAM_NAME) || true


.PHONY: print-make-variables
print-make-variables:
	@$(foreach V,$(sort $(.VARIABLES)), \
	   $(if $(filter-out environment% default automatic, \
	   $(origin $V)),$(warning $V=$($V) ($(value $V)))))


.PHONY: help
help:
	@echo "Build $(PROGRAM_NAME) version $(BUILD_VERSION)-$(BUILD_ITERATION)".
	@echo "All targets:"
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs
