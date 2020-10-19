# Development Setup

This document describes the recommended development setup for the [KubeVirt provider extension](https://github.com/gardener/gardener-extension-provider-kubevirt). Following the guidelines presented here would allow you to test the full Gardener reconciliation and deletion flows with the KubeVirt provider extension and the [KubeVirt MCM extension](https://github.com/gardener/machine-controller-manager-provider-kubevirt). 

In this setup, only Gardener itself is running in your local development cluster. All other components, as well as KubeVirt VMs, are deployed and run on external clusters, which avoids high CPU and memory load on your local laptop. 

## Prerequisites

Follow the steps outlined in [Setting up a local development environment](https://github.com/gardener/gardener/blob/master/docs/development/local_setup.md) for Gardener in order to install all needed prerequisites and enable running `gardener-apiserver`, `gardener-controller-manager`, and `gardenlet` locally. You can use either minikube, kind, or the nodeless cluster as your local development cluster.

Before continuing, copy all files from `docs/development/manifests` and `docs/development/scripts` to your `dev` directory and adapt them as needed. The sections that follow assume that you have already done this and all needed manifests and scripts can be found in your `dev` directory.

## Creating the ControllerRegistrations

Before you register seeds or create shoots, you need to register all needed extensions using `ControllerRegistration` resources. The easiest way to manage `ControllerRegistrations` is via [gem](https://github.com/gardener/gem). 

After installing `gem`, create a `requirements.yaml` file similar to [requirements.yaml](manifests/requirements.yaml). The example file contains only the extensions needed for the development setup described here, but you could add any other Gardener extensions you may need. 

In your `requirements.yaml` file you can refer to a released extension version, or to a revision (commit) from a Gardener repo or your fork of it. This version or revision is used to find the correct `controller-registration.yaml` file for the extension.

You can generate or update the `controller-registrations.yaml` file out of your `requirements.yaml` file by running:

```shell script
gem ensure --requirements dev/requirements.yaml --controller-registrations dev/controller-registrations.yaml
```

After generating or updating the `controller-registrations.yaml` file, review it and make sure all versions are the ones you want to use for your tests. For example, if you are working on a PR for the KubeVirt provider extension, in addition to specifying the revision in your fork in `requirements.yaml`, you may need to change the version from `0.1.0-dev` to something unique to you or your PR, e.g. `0.1.0-dev-johndoe`. You can also add `pullPolicy: Always` to ensure that if you push a new extension image with that version and delete the corresponding pod, the new image will always be pulled when the pod is recreated.

Once you are satisfied with your controller registrations, apply the `controller-registrations.yaml` to your local Gardener:

```shell script
kubectl apply -f dev/controller-registrations.yaml
```

## Registering the Seed Cluster

Create or choose an external cluster, different from your local development cluster, to register as seed in your local Gardener. This can be any cluster and it can be the same or different from your [provider cluster](#creating-the-provider-cluster). It is recommended to use a different cluster to avoid confusion between the two. If you want to use your provider cluster as seed, first create it as described below and then return to this step.

To register your cluster as a seed, create the secret containing the kubeconfig for your seed cluster, the secret containing the credentials for your cloud provider (e.g. GCP), and the seed resource itself. See the following files as examples:

* [secret-gcp1-kubeconfig.yaml](manifests/secret-gcp1-kubeconfig.yaml)
* [secret-seed-operator-gcp.yaml](manifests/secret-seed-operator-gcp.yaml)
* [seed-gcp1.yaml](manifests/seed-gcp1.yaml)

```shell script
kubectl apply -f dev/secret-gcp1-kubeconfig.yaml
kubectl apply -f dev/secret-seed-operator-gcp.yaml
kubectl apply -f dev/seed-gcp1.yaml
```

## Creating the Project

At this point, you should create a `dev` project in your local Gardener.

Create the project resource for your local `dev` project, see [project-dev](manifests/project-dev.yaml) as an example.

```shell script
kubectl apply -f dev/project-dev.yaml
```

## Creating the DNS Domain Secrets

At this point, you should create the domain secrets used by the DNS extension.

If you want to use an external DNS provider (e.g. route53), create default and internal domain secrets similar to [secret-default-domain.yaml](manifests/secret-default-domain.yaml) and [secret-internal-domain.yaml](manifests/secret-internal-domain.yaml).

```shell script
kubectl apply -f dev/secret-default-domain.yaml
kubectl apply -f dev/secret-internal-domain.yaml
```

Alternatively, if you don't want to use an external DNS provider and use `nip.io` addresses instead, create just an internal domain secret similar to [10-secret-internal-domain-unmanaged.yaml](https://github.com/gardener/gardener/blob/master/example/10-secret-internal-domain-unmanaged.yaml). For more information, see [Prepare the Gardener](https://github.com/gardener/gardener/blob/master/docs/development/local_setup.md#prepare-the-gardener).
   
## Creating the Provider Cluster

Create or choose an external cluster, different from your local development cluster, to use as a provider cluster. The only requirement to this cluster is that virtualization extensions are supported on its nodes. You can check if this is the case as described in [Easy install using Cloud Providers](https://kubevirt.io/pages/cloud.html), by  executing the command `egrep 'svm|vmx' /proc/cpuinfo` and checking for non-empty output.

### Creating an OS Image with Nested Virtualization Enabled

Before you can create such a cluster, you need to ensure that nested virtualizaton is enabled for its instances by using an appropriate OS image. To create such an image in GCP, follow the steps described in [Enabling nested virtualization for VM instances](https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances). For example, to create a custom Ubuntu image with nested virtualizaton enabled based on Ubuntu 18.04, execute the following commands:

```
gcloud compute disks create ubuntu-disk1 
  --image-project ubuntu-os-cloud \
  --image ubuntu-1804-bionic-v20200916 \
  --zone us-central1-b
gcloud compute images create ubuntu-1804-bionic-v20200916-vmx-enabled \
  --source-disk ubuntu-disk1 \
  --source-disk-zone us-central1-b \
  --licenses "https://compute.googleapis.com/compute/v1/projects/vm-options/global/licenses/enable-vmx"
gcloud compute images list | grep ubuntu
```

Once the image has been created, to create the provider cluster, you could use any Kubernetes provisioning tool, including of course Gardener itself, to create a cluster using this image. 

### Creating the Provider Cluster Using Gardener

To create the provider cluster using Gardener, simply create a shoot in the seed you registered previously using a custom GCP cloud profile that contains the above image, such as [cloudprofile-gcp.yaml](manifests/cloudprofile-gcp.yaml). To do this, follow these steps:

1. Create the custom GCP cloud profile, for example [cloudprofile-gcp.yaml](manifests/cloudprofile-gcp.yaml).

    ```shell script
    kubectl apply -f dev/cloudprofile-gcp.yaml
    ```

2. Create the shoot secret binding, you could bind to the `seed-operator-gcp` secret you created previously for your seed, see [secretbinding-shoot-operator-gcp.yaml](manifests/secretbinding-shoot-operator-gcp.yaml) as an example.

    ```shell script
    kubectl apply -f dev/secretbinding-shoot-operator-gcp.yaml
    ```

3. Create the GCP shoot itself. See [shoot-gcp-vmx.yaml](manifests/shoot-gcp-vmx.yaml) as an example. Note that this shoot should use the image with name `ubuntu` and version `18.4.20200916-vmx` from the custom GCP cloud profile you created previously. Also, please rename the shoot to contain an unique prefix such as your github username, e.g. `johndoe-gcp-vmx`, to avoid naming conflicts in GCP.

    ```shell script
    kubectl apply -f dev/shoot-gcp-vmx.yaml
    ```

    During the reconciliation by your local `gardenlet`, you may want to connect to the seed to monitor the shoot namespace `shoot--dev--<prefix>-gcp-vmx`.   

4. Once the shoot is successfully reconciled by your local `gardenlet`, get its kubeconfig by executing:

    ```shell script
    kubectl get secret <prefix>-gcp-vmx.kubeconfig -n garden-dev -o jsonpath={.data.kubeconfig} | base64 -d > dev/kubeconfig-gcp-vmx.yaml
    ```
   
### Installing KubeVirt, CDI, and Multus in the Provider Cluster

Once the provider cluster has been created (with Gardener or any other provisioning tool), you should install KubeVirt, CDI, and optionally Multus in it so that it can serve its purpose as a provider cluster. 
   
1. Install KubeVirt and CDI in this cluster by executing the [install-kubevirt.sh](../hack/kubevirt/install-kubevirt.sh) script:

    ```shell script
    export KUBECONFIG=dev/kubeconfig-gcp-vmx.yaml
    hack/kubevirt/install-kubevirt.sh
    ```
   
2. Optionally, to use networking features, install [Multus CNI](https://intel.github.io/multus-cni/doc/quickstart.html) as described in its documentation, or by applying the provided [multus.yaml](../hack/kubevirt/multus.yaml) manifest.

    ```shell script
    export KUBECONFIG=dev/kubeconfig-gcp-vmx.yaml
    kubectl apply -f hack/kubevirt/multus.yaml
    ```

    **Note:** In order to use any additional CNI plugins, the plugin binaries must be present in the `/opt/cni/bin` directory of the provider cluster nodes. For testing purposes, they can be installed manually by downloading a [containernetworking/plugins](https://github.com/containernetworking/plugins) release and copying the needed plugins to the `/opt/cni/bin` directory of each provider cluster node.

## Testing the Gardener Reconciliation Flow

To test the Gardener reconciliation flow with the KubeVirt provider extensions, create the KubeVirt shoot cluster in your local `dev` project, by following these steps:

1. Create the KubeVirt cloud profile, for example [cloudprofile-kubevirt.yaml](manifests/cloudprofile-kubevirt.yaml).

    ```shell script
    kubectl apply -f dev/cloudprofile-kubevirt.yaml
    ```

   **Note:** The example cloud profile is intentionally rather simple and does not take advantage of some of the features supported by the KubeVirt provider extension. To test these features, modify the cloud profile manifest accordingly. For more information, see [Using the KubeVirt provider extension with Gardener as operator](usage-as-operator.md).
   
2. Create the shoot secret and secret binding. You should create a secret containing the kubeconfig for your provider cluster, and a corresponding secret binding:

    ```shell script
    kubectl create secret generic kubevirt-credentials -n garden-dev --from-file=kubeconfig=dev/kubeconfig-gcp-vmx.yaml
    kubectl apply -f dev/secretbinding-kubevirt-credentials.yaml
    ```

3. Create the KubeVirt shoot itself. See [shoot-kubevirt.yaml](manifests/shoot-kubevirt.yaml) as an example. Note that the nodes CIDR for this shoot must be the same range as the pods CIDR of your provider cluster.

    ```shell script
    kubectl apply -f dev/shoot-kubevirt.yaml
    ```
   
   **Note:** The example shoot is intentionally very simple and does not take advantage of many of the features supported by the KubeVirt provider extension. To test these features, modify the shoot manifest accordingly. For more information, see [Using the KubeVirt provider extension with Gardener as end-user](usage-as-end-user.md).
   
4. During the shoot reconciliation by your local `gardenlet`, you may want to:

    * Monitor the `gardenlet` logs in your local console where `gardenlet` is running.
    * Connect to the seed to monitor the shoot namespace `shoot--dev--kubevirt` and the logs of the KubeVirt provider extension in the `extension-provider-kubevirt-*` namespace.
    * Connect to the provider cluster to monitor the `default` namespace where VMs and VMIs are being created.

5. Once the shoot has been successfully reconciled, get its kubeconfig by executing:

    ```shell script
    kubectl get secret kubevirt.kubeconfig -n garden-dev -o jsonpath={.data.kubeconfig} | base64 -d > dev/kubeconfig-kubevirt.yaml
    ```
   
    At this point, you may want to connect to the KubeVirt shoot and check if it's usable.
    
## Testing the Gardener Deletion Flow

To test the Gardener deletion flow with the KubeVirt provider extensions, delete the KubeVirt shoot cluster in your local `dev` project, by following these steps:

1. Delete the KubeVirt shoot itself using the [delete](https://github.com/gardener/gardener/blob/master/hack/usage/delete) script.

    ```shell script
    kubectl annotate shoot kubevirt -n garden-dev confirmation.gardener.cloud/deletion=1
    kubectl delete shoot kubevirt -n garden-dev
    ```
   
2. During the shoot deletion by your local `gardenlet`, you may want to:

    * Monitor the `gardenlet` logs in your local console where `gardenlet` is running.
    * Connect to the seed to monitor the shoot namespace `shoot--dev--kubevirt` and the logs of the KubeVirt provider extension in the `extension-provider-kubevirt-*` namespace.
    * Connect to the provider cluster to monitor the `default` namespace where VMs and VMIs are being created.
