#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail


source $(dirname ${0})/../kind/common.sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

GIT_SERVER_BASE_URL="http://gitops:testtest@git.local.gardener.cloud:6080"

ensure_glk_configuration() {
  echo "⚙️  Ensuring GLK configuration"
  cp "$SCRIPT_DIR/landscapekitconfiguration.yaml" "${GLK_CONFIG_PATH}"
}

clone_or_update_repo() {
  repoName=$1
  destSubDir=$2

  repoUrl=$GIT_SERVER_BASE_URL/gitops/${repoName}.git
  destDir="${WORK_DIR}/${destSubDir}"

  if [ -d "${destDir}/.git" ]; then
    git -C ${destDir} pull
  else
    git clone ${repoUrl} "${destDir}"
  fi

  pushd "${destDir}"
  git config user.name 'Gitops'
  git config user.email 'gitops@gardener'
  popd > /dev/null
}

checkout_base_repo() {
  echo "📥 Checking out base repository"
  clone_or_update_repo base base
}

generate_base() {
  echo "🌱 Generating base"
  glk generate base -c "${GLK_CONFIG_PATH}" "${GLK_BASE_REPO_PATH}"

  local glk_dev_image=$(cat $SCRIPT_DIR/../glk-dev-image)
  if [ -z "$glk_dev_image" ]; then
    echo "GLK_DEV_IMAGE is empty. Please build a dev version with Skaffold before setting up the repositories."
    exit 1
  fi

  local glk_dev_image_base=${glk_dev_image%:*}
  local glk_dev_image_version=${glk_dev_image##*:}

  local workflows_path="${GLK_BASE_REPO_PATH}/.github/workflows"
  cp "$SCRIPT_DIR/workflow-pr-post-change.yaml" "$workflows_path"
  sed -i "s|<COMMAND>|base|g" "$workflows_path/workflow-pr-post-change.yaml"
  sed -i "s|<BASE-PATH>|./|g" "$workflows_path/workflow-pr-post-change.yaml"
  sed -i "s|<IMAGE-BASE>|$glk_dev_image_base|g" "$workflows_path/workflow-pr-post-change.yaml"

  cp "$SCRIPT_DIR/components.yaml" "${GLK_BASE_REPO_PATH}/components.yaml"
  sed -i "s|<DEV-VERSION>|$glk_dev_image_version|g" "${GLK_BASE_REPO_PATH}/components.yaml"

  cd "${GLK_BASE_REPO_PATH}"
  git add .
  git commit -m "Generate base" || echo "No changes to commit"
  git push
}

checkout_landscape_repo() {
  echo "📥 Checking out test landscape repository"
  clone_or_update_repo test-landscape test-landscape
}

generate_landscape() {
  echo "🌱 Generating test landscape"
  glk generate landscape -c "${GLK_CONFIG_PATH}" "${GLK_LANDSCAPE_REPO_PATH}"

  local glk_dev_image=$(cat $SCRIPT_DIR/../glk-dev-image)
  if [ -z "$glk_dev_image" ]; then
    echo "GLK_DEV_IMAGE is empty. Please build a dev version with Skaffold before setting up the repositories."
    exit 1
  fi

  local glk_dev_image_base=${glk_dev_image%:*}

  local workflows_path="${GLK_LANDSCAPE_REPO_PATH}/.github/workflows"
  cp "$SCRIPT_DIR/workflow-pr-post-change.yaml" "$workflows_path"
  sed -i "s|<COMMAND>|landscape|g" "$workflows_path/workflow-pr-post-change.yaml"
  sed -i "s|<BASE-PATH>|./base/|g" "$workflows_path/workflow-pr-post-change.yaml"
  sed -i "s|<IMAGE-BASE>|$glk_dev_image_base|g" "$workflows_path/workflow-pr-post-change.yaml"

  cd "${GLK_LANDSCAPE_REPO_PATH}"
  git add .
  git commit -m "Generate test landscape" || echo "No changes to commit"
  git push
}

ensure_base_as_submodule() {
  echo "🔗 Ensuring base is a submodule of test-landscape"
  cd "${GLK_LANDSCAPE_REPO_PATH}"

  if [ ! -f .gitmodules ] || ! grep -q "\[submodule \"base\"\]" .gitmodules; then
    git submodule add $GIT_SERVER_BASE_URL/gitops/base.git base
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

checkout_base_repo
ensure_glk_configuration
generate_base
checkout_landscape_repo
ensure_base_as_submodule
generate_landscape
