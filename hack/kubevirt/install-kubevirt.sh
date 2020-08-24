#!/usr/bin/env bash

set -euxo pipefail

# shellcheck source=hack/common.sh
source "$(dirname "$0")"/common.sh

kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/"${CDI_VERSION}"/cdi-operator.yaml
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/"${CDI_VERSION}"/cdi-cr.yaml

kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/"${KUBEVIRT_VERSION}"/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/"${KUBEVIRT_VERSION}"/kubevirt-cr.yaml

cat <<EOF | kubectl create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubevirt-config
  namespace: kubevirt
  labels:
    kubevirt.io: ""
data:
  feature-gates: "DataVolumes"
EOF
