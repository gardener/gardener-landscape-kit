#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

SCRIPT_DIR=$(dirname ${0})
source $(dirname ${0})/common.sh

GIT_SERVER_BASE_URL="http://test:testtest@git.local.gardener.cloud:6080"

ensure_glk_configuration() {
  echo "⚙️  Ensuring GLK configuration"
  cp "$SCRIPT_DIR/landscapekitconfiguration.yaml" "${WORK_DIR}/landscapekitconfiguration.yaml"
}

clone_or_update_repo() {
  repoName=$1
  destSubDir=$2

  repoUrl=$GIT_SERVER_BASE_URL/test/${repoName}.git
  destDir="${WORK_DIR}/${destSubDir}"

  if [ -d "${destDir}/.git" ]; then
    git -C ${destDir} pull
  else
    git clone ${repoUrl} "${destDir}"
  fi
}

generate_base() {
  echo "🌱 Generating base"
  clone_or_update_repo base base

  glk generate base -c "${WORK_DIR}/landscapekitconfiguration.yaml" "${WORK_DIR}/base"

  cd "${WORK_DIR}/base"
  git add .
  git commit -m "Generate base" || echo "No changes to commit"
  git push
}

generate_landscape() {
  echo "🌱 Generating test landscape"
  clone_or_update_repo test-landscape test-landscape

  glk generate landscape -c "${WORK_DIR}/landscapekitconfiguration.yaml" "${WORK_DIR}/test-landscape"

  cd "${WORK_DIR}/test-landscape"
  git add .
  git commit -m "Generate test landscape" || echo "No changes to commit"
  git push
}

ensure_base_as_submodule() {
  echo "🔗 Ensuring base is a submodule of test-landscape"
  cd "${WORK_DIR}/test-landscape"

  if [ ! -f .gitmodules ] || ! grep -q "\[submodule \"base\"\]" .gitmodules; then
    git submodule add $GIT_SERVER_BASE_URL/test/base.git base
    git add .gitmodules base
    git commit -m "Add base as submodule"
    git push
  else
    echo "Base is already a submodule"
    git submodule update --remote --rebase base
    git add base
    git commit -m "Update base submodule" || echo "No changes to commit"
    git push
  fi
  git submodule update --init
}

ensure_glk_configuration
generate_base
generate_landscape
ensure_base_as_submodule
