# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

GLK_PRETTIFY        := $(TOOLS_BIN_DIR)/prettify

FLUX_CLI_VERSION ?= $(shell grep -A3 'fluxCLI:' $(REPO_ROOT)/componentvector/components.yaml | sed -n 's/.*tag: //p')
GLK_PRETTIFY_VERSION = $(shell git rev-parse HEAD)

FLUX_CLI ?= $(TOOLS_DIR)/bin/$(SYSTEM_NAME)-$(SYSTEM_ARCH)/flux
.PHONY: flux-cli
flux-cli: $(FLUX_CLI)
$(FLUX_CLI): $(TOOLS_DIR) $(call tool_version_file,$(FLUX_CLI),$(FLUX_CLI_VERSION))
	@mkdir -p $(dir $(FLUX_CLI))
	@printf '#!/usr/bin/env bash\nset -e\ndocker run --rm -v "$(REPO_ROOT):$(REPO_ROOT)" ghcr.io/fluxcd/flux-cli:$(FLUX_CLI_VERSION) "$$@"\n' > $(FLUX_CLI)
	@chmod +x $(FLUX_CLI)

$(GLK_PRETTIFY): $(call tool_version_file,$(GLK_PRETTIFY),$(GLK_PRETTIFY_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install ./hack/tools/prettify
