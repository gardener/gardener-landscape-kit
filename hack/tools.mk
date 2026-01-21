# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

GLK_GLK             := $(TOOLS_BIN_DIR)/gardener-landscape-kit
GLK_PRETTIFY        := $(TOOLS_BIN_DIR)/prettify

# renovate: datasource=github-releases depName=fluxcd/flux2
FLUX_CLI_VERSION ?= v2.7.5
GLK_GLK_VERSION = $(shell git rev-parse HEAD)
GLK_PRETTIFY_VERSION = $(shell git rev-parse HEAD)

FLUX_CLI ?= $(TOOLS_DIR)/bin/$(SYSTEM_NAME)-$(SYSTEM_ARCH)/flux
.PHONY: flux-cli
flux-cli: $(FLUX_CLI)
$(FLUX_CLI): $(TOOLS_DIR) $(call tool_version_file,$(FLUX_CLI),$(FLUX_CLI_VERSION))
	curl -Lo $(FLUX_CLI).tar.gz https://github.com/fluxcd/flux2/releases/download/$(FLUX_CLI_VERSION)/flux_$(FLUX_CLI_VERSION:v%=%)_$(SYSTEM_NAME)_$(SYSTEM_ARCH).tar.gz
	tar -zxvf $(FLUX_CLI).tar.gz -C $(TOOLS_DIR)/bin/$(SYSTEM_NAME)-$(SYSTEM_ARCH) flux
	touch $(FLUX_CLI) && chmod +x $(FLUX_CLI) && rm $(FLUX_CLI).tar.gz

$(GLK_GLK): $(call tool_version_file,$(GLK_GLK),$(GLK_GLK_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install ./cmd/gardener-landscape-kit

$(GLK_PRETTIFY): $(call tool_version_file,$(GLK_PRETTIFY),$(GLK_PRETTIFY_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install ./hack/tools/prettify
