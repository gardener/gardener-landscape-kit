# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

NAME                 := gardener-landscape-kit
VERSION              := $(shell cat VERSION)
EFFECTIVE_VERSION    := $(VERSION)-$(shell git rev-parse HEAD)
REPO_ROOT            := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR             := $(REPO_ROOT)/hack
ENSURE_GARDENER_MOD  := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
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
BUILD_PACKAGES    ?= ./cmd/...

.PHONY: build
build:
	@LD_FLAGS=$(LD_FLAGS) EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) bash $(GARDENER_HACK_DIR)/build.sh -o $(BUILD_OUTPUT_FILE) $(BUILD_PACKAGES)

.PHONY: install
install:
	@LD_FLAGS=$(LD_FLAGS) bash $(GARDENER_HACK_DIR)/install.sh ./cmd/...

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: generate
generate: $(GEN_CRD_API_REFERENCE_DOCS) $(VGOPATH)
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(REPO_ROOT)/hack/update-codegen.sh

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: clean
clean:
	@bash $(GARDENER_HACK_DIR)/clean.sh ./pkg/...

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

.PHONY: verify-extended
verify-extended: check-generate check format test-cov sast-report

generate-ocm-testdata:
	@go run ./hack/tools/ocm-testdata-generator -config $(REPO_ROOT)/test/ocm/testdata/config.yaml