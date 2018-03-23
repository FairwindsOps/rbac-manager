#!/usr/bin/env bash

# This script is useful when managing service accounts
# Run ./generate_config.sh default my-service-account
# to get a valid KUBECONFIG file for kubectl to auth as the
# specified service account

set -eo pipefail

kc() {
  kubectl -n "${namespace}" $@
}

namespace="$1"
serviceaccount="$2"

sa_secret_name=$(kc get serviceaccount "${serviceaccount}" -o 'jsonpath={.secrets[0].name}')

context_name="$(kubectl config current-context)"
cluster_name="$(kubectl config view -o "jsonpath={.contexts[?(@.name==\"${context_name}\")].context.cluster}")"
server_name="$(kubectl config view -o "jsonpath={.clusters[?(@.name==\"${cluster_name}\")].cluster.server}")"
cacert="$(kc get secret "${sa_secret_name}" -o "jsonpath={.data.ca\.crt}" | base64 --decode)"
token="$(kubectl get secret "${sa_secret_name}" -o "jsonpath={.data.token}" | base64 --decode)"

export KUBECONFIG="$(mktemp)"
kubectl config set-credentials "${serviceaccount}" --token="${token}" >/dev/null
ca_crt="$(mktemp)"; echo "${cacert}" > ${ca_crt}
kubectl config set-cluster "${cluster_name}" --server="${server_name}" --certificate-authority="$ca_crt" --embed-certs >/dev/null
kubectl config set-context "${cluster_name}" --cluster="${cluster_name}" --user="${serviceaccount}" >/dev/null
kubectl config use-context "${cluster_name}" >/dev/null

KUBECONFIG_DATA=$(cat "${KUBECONFIG}")
rm ${KUBECONFIG}
echo "${KUBECONFIG_DATA}"
