#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

source $(dirname ${0})/common.sh
SCRIPT_DIR=$(dirname ${0})

clusterNameSuffix=$1

glkSuffix="glk"
if [[ $clusterNameSuffix == "single" ]]; then
  glkSuffix="single"
fi

clusterName="$GLK_KIND_CLUSTER_PREFIX-$clusterNameSuffix"

install_metallb() {
  echo "🚀 install metal loadbalance on kind cluster $clusterName"
  # install metal loadbalancer (see https://kind.sigs.k8s.io/docs/user/loadbalancer/)
  kubectl apply -k "$REPO_ROOT/dev-setup/kind/metallb" --server-side
  kubectl wait --namespace metallb-system --for=condition=available deployment --selector=app=metallb --timeout=90s

  kindIPAM=$(docker network inspect -f '{{range .IPAM.Config}}{{.Subnet}} {{end}}' kind)
  if [[ "$kindIPAM" =~ ([0-9]+\.[0-9]+)(".0.0/24 ") ]]; then
    cidrPrefix=${BASH_REMATCH[1]}
    cidr="$cidrPrefix.0.0/24"
    echo "kind network cidr: $cidr"
  else
    echo "cannot extract IPv4 CIDR from '$kindIPAM'"
  fi

  start_range=$cidrPrefix.255.100
  end_range=$cidrPrefix.255.254

  sed -e "s/#range_start/$start_range/g" -e "s/#range_end/$end_range/g" "$REPO_ROOT/dev-setup/kind/metallb/ipaddresspool.yaml.template" | \
    kubectl apply -f -
}

# setup_kind_network is similar to kind's network creation logic, ref https://github.com/kubernetes-sigs/kind/blob/23d2ac0e9c41028fa252dd1340411d70d46e2fd4/pkg/cluster/internal/providers/docker/network.go#L50
# In addition to kind's logic, we ensure stable CIDRs that we can rely on in our local setup manifests and code.
setup_kind_network() {
  # check if network already exists
  local existing_network_id
  existing_network_id="$(docker network list --filter=name=^kind$ --format='{{.ID}}')"

  if [ -n "$existing_network_id" ] ; then
    # ensure the network is configured correctly
    local network network_options network_ipam expected_network_ipam
    network="$(docker network inspect $existing_network_id | yq '.[]')"
    network_options="$(echo "$network" | yq '.EnableIPv6 + "," + .Options["com.docker.network.bridge.enable_ip_masquerade"]')"
    network_ipam="$(echo "$network" | yq '.IPAM.Config' -o=json -I=0 | sed -E 's/"IPRange":"",//g')"
    expected_network_ipam='[{"Subnet":"172.18.0.0/24","Gateway":"172.18.0.1"},{"Subnet":"fd00:10::/64","Gateway":"fd00:10::1"}]'

    if [ "$network_options" = 'true,true' ] && [ "$network_ipam" = "$expected_network_ipam" ] ; then
      # kind network is already configured correctly, nothing to do
      return 0
    else
      echo "kind network is not configured correctly for local gardener setup, recreating network with correct configuration..."
      docker network rm $existing_network_id
    fi
  fi

  # (re-)create kind network with expected settings
  docker network create kind --driver=bridge \
    --subnet 172.18.0.0/24 --gateway 172.18.0.1 \
    --ipv6 --subnet fd00:10::/64 --gateway fd00:10::1 \
    --opt com.docker.network.bridge.enable_ip_masquerade=true
}

build_kind_node_image() {
  echo "### building kind node image"

  docker build -t glk-kind-node:latest -f $SCRIPT_DIR/node/Dockerfile $SCRIPT_DIR/node
}

create_kind_cluster() {
  echo "🚀 creating kind cluster $clusterName"

  mkdir -p ${REPO_ROOT}/dev
  export KUBECONFIG=${REPO_ROOT}/dev/kind-$clusterName-kubeconfig.yaml
  # only create cluster if not existing
  kind get clusters | grep $clusterName &> /dev/null || \
    kind create cluster \
      --name $clusterName \
      --config <(helm template "$SCRIPT_DIR/clusters/base")
}

setup_kind_network
build_kind_node_image
create_kind_cluster
install_metallb

echo "ℹ️ To access $clusterName cluster, use:"
echo "export KUBECONFIG=$KUBECONFIG"
echo ""
