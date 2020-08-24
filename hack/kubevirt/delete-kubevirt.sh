#!/usr/bin/env bash

set -euxo pipefail

# shellcheck source=hack/common.sh
source "$(dirname "$0")"/common.sh

kubectl delete --ignore-not-found configmap kubevirt-config -n kubevirt

kubectl delete --ignore-not-found -f https://github.com/kubevirt/kubevirt/releases/download/"${KUBEVIRT_VERSION}"/kubevirt-cr.yaml
kubectl delete --ignore-not-found -f https://github.com/kubevirt/kubevirt/releases/download/"${KUBEVIRT_VERSION}"/kubevirt-operator.yaml

kubectl delete --ignore-not-found -f https://github.com/kubevirt/containerized-data-importer/releases/download/"${CDI_VERSION}"/cdi-cr.yaml
kubectl delete --ignore-not-found -f https://github.com/kubevirt/containerized-data-importer/releases/download/"${CDI_VERSION}"/cdi-operator.yaml
