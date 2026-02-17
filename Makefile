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
IMAGE_REGISTRY       ?= europe-docker.pkg.dev/gardener-project/snapshots/gardener/gardener-landscape-kit
TARGET_PLATFORMS     ?= linux/$(shell go env GOARCH)

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk
include $(HACK_DIR)/tools.mk

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

.PHONY: docker-images
docker-images:
	@echo "Building docker images with version and tag $(EFFECTIVE_VERSION) for target platforms $(TARGET_PLATFORMS)"
	@docker build --build-arg EFFECTIVE_VERSION=$(EFFECTIVE_VERSION) --platform $(TARGET_PLATFORMS) -t $(IMAGE_REGISTRY):$(EFFECTIVE_VERSION) -t $(IMAGE_REGISTRY):latest -f Dockerfile --target gardener-landscape-kit .

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy
	@cd $(HACK_DIR)/tools/mod && go mod tidy

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) MODE=$(MODE) bash $(HACK_DIR)/format.sh ./cmd ./pkg

tools-for-generate: $(GEN_CRD_API_REFERENCE_DOCS)
	@go mod download

.PHONY: generate
generate: tools-for-generate $(GOIMPORTS) $(FLUX_CLI) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./componentvector/... ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(HACK_DIR)/update-codegen.sh
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(HACK_DIR)/update-github-templates.sh
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(HACK_DIR)/generate-renovate-ignore-deps.sh
	$(MAKE) format

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
sast:
	@HACK_DIR=$(HACK_DIR) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(HACK_DIR)/sast.sh --exclude-dirs hack,dev

.PHONY: sast-report
sast-report:
	@HACK_DIR=$(HACK_DIR) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(HACK_DIR)/sast.sh --exclude-dirs hack,dev --gosec-report true

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

.PHONY: generate-ocm-testdata
generate-ocm-testdata:
	@go run $(HACK_DIR)/tools/ocm-testdata-generator -config $(REPO_ROOT)/pkg/ocm/components/testdata/config.yaml

.PHONY: git-server-up
git-server-up:
	@bash $(REPO_ROOT)/dev-setup/git-server/git-server-up.sh

.PHONY: git-server-down
git-server-down:
	@bash $(REPO_ROOT)/dev-setup/git-server/git-server-down.sh

.PHONY: git-server-cleanup # cleanup git server data
git-server-cleanup: git-server-down $(YQ)
	@rm -rf $(REPO_ROOT)/dev/git-server/data

.PHONY: registry-up
registry-up:
	@$(REPO_ROOT)/dev-setup/registry/registry-up.sh

.PHONY: registry-down
registry-down:
	@$(REPO_ROOT)/dev-setup/registry/registry-down.sh

.PHONY: kind-up ## create single kind cluster for hosting glk and runtime
kind-up: $(KIND) $(KUBECTL) $(HELM)
	@$(REPO_ROOT)/dev-setup/kind/kind-create-cluster.sh single
	@$(MAKE) registry-up git-server-up

.PHONY: kind-down
kind-down: git-server-down registry-down $(KIND) $(KUBECTL)
	@$(REPO_ROOT)/dev-setup/kind/kind-delete-cluster.sh single

#########################################
# E2E Tests                             #
#########################################

KIND_LOCAL_KUBECONFIG ?= $(REPO_ROOT)/dev/kind-glk-single-kubeconfig.yaml
PARALLEL_E2E_TESTS    ?= 5

ifndef ARTIFACTS
	export ARTIFACTS=/tmp/artifacts
endif

e2e-prepare: export BUILD_OUTPUT_FILE=$(TOOLS_BIN_DIR)

.PHONY: e2e-prepare
e2e-prepare: build $(SKAFFOLD) $(HELM) $(KUBECTL) $(YQ) $(GLK_PRETTIFY)
	@$(REPO_ROOT)/dev-setup/kind/generate-repos.sh
	@$(REPO_ROOT)/dev-setup/kind/deploy-flux.sh
	@$(REPO_ROOT)/dev-setup/kind/prepare-garden.sh
	@$(REPO_ROOT)/dev-setup/kind/build-and-add-provider-local.sh

.PHONY: ci-e2e-kind
ci-e2e-kind:
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(HACK_DIR)/ci-e2e-kind.sh

.PHONY: test-e2e-local
test-e2e-local: $(GINKGO)
	KUBECONFIG=$(KIND_LOCAL_KUBECONFIG) $(REPO_ROOT)/dev/gardener/hack/test-e2e-local.sh --procs=$(PARALLEL_E2E_TESTS) --label-filter="default" ./test/e2e/...
