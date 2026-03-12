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

SUDO=""
if [[ "$(id -u)" != "0" ]]; then
  SUDO="sudo "
fi

# The local registry needs to be available from the host for pushing and from the containers for pulling.
# On the host, we bind the registry to localhost (see infra/docker-compose.yaml), because 127.0.0.1 and ::1
# are configured as HTTP-only (insecure-registries) by default in Docker, which allows `docker push` without
# changing the Docker daemon config.
# From within the containers (e.g., the kind nodes), the registry domain is resolved via Docker's built-in
# DNS server to the IP of the registry container because of the host alias configured in docker compose.
#
# We could also bind the registry to an 172.18.255.* address similar to bind9 to make the registry reachable
# from the host and containers via the same IP. This would be cleaner, because we wouldn't need to add entries
# to /etc/hosts for the registry domain and would resolve the domain from the host via bind9 just as all other
# domains.
# However, this would require changing the Docker daemon config to set registry.local.gardener.cloud as an
# insecure registry to allow pushing to it from the host. The insecureRegistries settings in the skaffold config
# doesn't apply here, because skaffold uses the Docker daemon/CLI under the hood for pushing images, which only
# considers the Docker daemon's registry configuration.
ensure_local_registry_hosts() {
  local host="glk-registry.local.gardener.cloud"

  for ip in 127.0.0.1 ::1 ; do
    if ! grep -Eq "^$ip $host$" /etc/hosts; then
      echo "> Adding entry '$ip $host' to /etc/hosts..."
      echo "$ip $host" | ${SUDO}tee -a /etc/hosts
    else
      echo "> /etc/hosts already contains entry '$ip $host', skipping..."
    fi
  done
}

setup_local_dns_resolver() {
  local dns_ip=172.18.255.54
  local dns_ipv6=fd00:ff::54

  # Special handling in CI: we don't have a fully-fledged systemd-resolved or similar in the CI environment, so we set
  # up dnsmasq as a local DNS resolver with conditional forwarding for the local.gardener.cloud domain to the local
  # setup's DNS server (bind9).
  # Setting bind9 as the nameserver in /etc/resolv.conf directly does not work, as bind9 itself forwards to the host's
  # nameservers configured in resolv.conf, creating a cyclic dependency. With dnsmasq however, we can configure it to
  # forward requests only for the local.gardener.cloud domain to the local setup's DNS server, and forward all other
  # requests to the default nameservers (the Prow cluster's coredns), which works fine.
  if [ -n "${CI:-}" -a -n "${ARTIFACTS:-}" ]; then
    mkdir -p /etc/dnsmasq.d/
    cp /etc/resolv.conf /etc/resolv-default.conf
    tee /etc/dnsmasq.d/gardener-local.conf <<EOF
# Force dnsmasq to listen ONLY on standard localhost and prevent it from scanning other interfaces/IPs.
# Without this, it ignores the server directive for local.gardener.cloud because the IP is bound to the loopback
# interface and assumes doing so would create an infinite loop.
listen-address=127.0.0.1
bind-interfaces

# Configure conditional forwarding for local.gardener.cloud but use the resolv.conf from Kubernetes (coredns) as
# upstream for all other requests, which is required for resolving the registry cache services in the Prow cluster.
server=/local.gardener.cloud/$dns_ip
resolv-file=/etc/resolv-default.conf

# Export dnsmasq logs to a file for debugging purposes
log-facility=/var/log/dnsmasq.log
log-queries
EOF

    service dnsmasq start || service dnsmasq restart

    echo "> Setting dnsmasq as nameserver in /etc/resolv.conf..."
    # /etc/resolv.conf is shared between all containers in the pod, i.e., it will also be used by the injected sidecar
    # containers (e.g., for uploading artifacts to GCS). Hence, we keep the previous nameservers as fallback if dnsmasq
    # is not working, but set dnsmasq as the first entry to ensure it is used as primary resolver for the test job.
    # We cannot use sed -i on the /etc/resolv.conf bind mount that Kubernetes adds, so we need to write to a temp file
    # and then overwrite the resolv.conf with the combined content.
    echo "nameserver 127.0.0.1" > /tmp/resolv.conf
    cat /etc/resolv.conf >> /tmp/resolv.conf
    cat /tmp/resolv.conf > /etc/resolv.conf
    rm /tmp/resolv.conf

    echo "> Content of /etc/resolv.conf after setting dnsmasq as nameserver"
    cat /etc/resolv.conf

    return 0
  fi

  if [[ "$OSTYPE" == "darwin"* ]]; then
    local desired_resolver_config="nameserver $dns_ip"
    if ! grep -q "$desired_resolver_config" /etc/resolver/local.gardener.cloud ; then
      echo "Configuring macOS to resolve the local.gardener.cloud zone using the local setup's DNS server"
      ${SUDO}mkdir -p /etc/resolver
      echo "$desired_resolver_config" | ${SUDO}tee /etc/resolver/local.gardener.cloud
    fi
  elif [[ "$OSTYPE" == "linux"* && -d /etc/systemd/resolved.conf.d ]]; then
    if ! grep -q "$dns_ip" /etc/systemd/resolved.conf.d/gardener-local.conf || ! grep -q "$dns_ipv6" /etc/systemd/resolved.conf.d/gardener-local.conf ; then
      echo "Configuring systemd-resolved to resolve the local.gardener.cloud zone using the local setup's DNS server"
      cat <<EOF | ${SUDO}tee /etc/systemd/resolved.conf.d/gardener-local.conf
[Resolve]
DNS=$dns_ip $dns_ipv6
Domains=~local.gardener.cloud
EOF
      echo "restarting systemd-resolved"
      ${SUDO}systemctl restart systemd-resolved
    fi
  elif ! nslookup -type=ns local.gardener.cloud >/dev/null 2>/dev/null ; then
    echo "Warning: Unknown OS. Make sure your host resolves the local.gardener.cloud zone using the local setup's DNS server at $dns_ip or $dns_ipv6 respectively."
    return 0
  fi
}

install_metallb() {
  echo "🚀 install metal loadbalancer on kind cluster $clusterName"
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
ensure_local_registry_hosts
setup_local_dns_resolver
create_kind_cluster
install_metallb

echo "ℹ️ To access $clusterName cluster, use:"
echo "export KUBECONFIG=$KUBECONFIG"
echo ""
