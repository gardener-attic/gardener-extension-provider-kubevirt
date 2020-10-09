# Using the KubeVirt provider extension with Gardener as operator

The [`core.gardener.cloud/v1beta1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration. The [`core.gardener.cloud/v1beta1.Seed` resource](https://github.com/gardener/gardener/blob/master/example/50-seed.yaml) is structured in a similar way. Additionally, it allows configuring settings for the backups of the main etcds' data of shoot clusters control planes running in this seed cluster.

This document explains what is necessary to configure for this provider extension.

## `CloudProfile` resource

In this section we are describing how the configuration for `CloudProfile`s looks like for KubeVirt and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating KubeVirt shoot clusters.

### `CloudProfileConfig`

The cloud profile configuration contains information about the machine images source URLs. You have to map every version that you specify in `.spec.machineImages[].versions` here so that the KubeVirt extension could find the source URL for every version you want to offer.

An example `CloudProfileConfig` for the KubeVirt extension looks as follows:

```yaml
apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: ubuntu
  versions:
  - version: "18.04"
    sourceURL: https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img
# machineTypes extend cloud profile's spec.machineType object to KubeVirt provider specific config
machineTypes:
# name is used as a reference to the machineType object
- name: standard-1  
  # limits is equivalent to resource limits of pod
  # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container
  limits:
    cpu: "2"
    memory: 8Gi
```

### Example `CloudProfile` manifest

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: kubevirt
spec:
  type: kubevirt
  kubernetes:
    versions:
    - version: "1.17.8"
    - version: "1.18.5"
  machineImages:
  - name: ubuntu
    versions:
    - version: "18.04"
  machineTypes:
  - name: standard-1
    cpu: "1"
    gpu: "0"
    memory: 4Gi
  volumeTypes:
  - name: default
    class: default
  regions:
  - name: europe-west1
    zones:
    - name: europe-west1-b
    - name: europe-west1-c
    - name: europe-west1-d
  providerConfig:
    apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    machineImages:
    - name: ubuntu
      versions:
      - version: "18.04"
        sourceURL: https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img
```

## `Seed` resource

This provider extension does not support any provider configuration for the `Seed`'s `.spec.provider.providerConfig` field.
