# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

NAME                 := gardener-landscape-kit
VERSION              := $(shell cat VERSION)
EFFECTIVE_VERSION    := $(VERSION)-$(shell git rev-parse HEAD)
REPO_ROOT            := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR             := $(REPO_ROOT)/hack
GARDENER_HACK_DIR    := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
LD_FLAGS             := "-w $(shell bash $(GARDENER_HACK_DIR)/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/VERSION $(NAME))"

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

#########################################
# Targets                               #
#########################################

BUILD_OUTPUT_FILE ?= ./dev/

.PHONY: build
build:
	@LD_FLAGS=$(LD_FLAGS) EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) bash $(GARDENER_HACK_DIR)/build.sh -o $(BUILD_OUTPUT_FILE) ./cmd/...

.PHONY: install
install:
	@LD_FLAGS=$(LD_FLAGS) bash $(GARDENER_HACK_DIR)/install.sh ./cmd/...

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...

.PHONY: sast
sast: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --exclude-dirs hack

.PHONY: sast-report
sast-report: $(GOSEC)
	@bash $(GARDENER_HACK_DIR)/sast.sh --exclude-dirs hack --gosec-report true

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-clean
test-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format test sast
