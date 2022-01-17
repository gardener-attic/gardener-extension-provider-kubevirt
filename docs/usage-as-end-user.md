# Using the KubeVirt provider extension with Gardener as end-user

The [`core.gardener.cloud/v1beta1.Shoot` resource](https://github.com/gardener/gardener/blob/master/example/90-shoot.yaml) declares a few fields that are meant to contain provider-specific configuration.

This document describes how this configuration looks like for KubeVirt and provides an example `Shoot` manifest with minimal configuration that you can use to create a KubeVirt shoot cluster (without the landscape-specific information such as cloud profile names, secret binding names, etc.).

## Provider Secret Data

Every shoot cluster references a `SecretBinding` which itself references a `Secret`, and this `Secret` contains the kubeconfig of your *KubeVirt provider cluster*. This cluster is the cluster where KubeVirt itself is installed, and that hosts the KubeVirt virtual machines used as shoot worker nodes. This `Secret` must look as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: provider-cluster-kubeconfig
  namespace: garden-dev
type: Opaque
data:
  kubeconfig: base64(kubeconfig)
```

### Permissions

All KubeVirt resources (`VirtualMachines`, `DataVolumes`, etc.) are created in the namespace of the current context of the above kubeconfig, that is `my-shoot` in the example below:

```yaml
...
current-context: provider-cluster
contexts:
- name: provider-cluster
  context:
    cluster: provider-cluster
    namespace: my-shoot
    user: provider-cluster-token
...
```

If no namespace is specified, the `default` namespace is assumed. You can use the same namespace for multiple shoots. The user specified in the `kubeconfig` must have permissions to read and write KubeVirt and Kubernetes core resources in this namespace.

## `InfrastructureConfig`

The infrastructure configuration can contain additional networks used by the shoot worker nodes. If this configuration is empty, all KubeVirt virtual machines used as shoot worker nodes use only the pod network of the provider cluster. 

An example `InfrastructureConfig` for the KubeVirt extension looks as follows:

```yaml
apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
kind: InfrastructureConfig
networks:
  sharedNetworks:
  # Reference to the network defined by the NetworkAttachmentDefinition default/net-conf
  - name: net-conf
    namespace: default
  tenantNetworks:
  - name: network-1
    # Configuration for the CNI plugins bridge and firewall
    config: |
      {
        "cniVersion": "0.4.0",
        "name": "bridge-firewall",
        "plugins": [
          {
            "type": "bridge",
            "isGateway": true,
            "isDefaultGateway": true,
            "ipMasq": true,
            "ipam": {
              "type": "host-local",
              "subnet": "10.100.0.0/16"
            }
          },
          {
            "type": "firewall"
          }
        ]
      }
    # Don't attach the pod network at all, instead use this network as default
    default: true
```

A non-empty infrastructure configuration can contain:

* References to pre-existing, *shared* networks that can be shared between multiple shoots. These networks must exist in the provider cluster prior to shoot creation.
* CNI configurations for *tenant* networks that are created, updated, and deleted together with the shoot. If one of these networks is marked as `default: true`, it becomes the default network instead of the pod network of the provider cluster. This can be used to achieve higher level of network isolation, since the networks of the different shoots can be isolated from each other, and in some cases better performance.

Both shared and tenant networks are maintained in the provider cluster via [Multus CNI](https://github.com/intel/multus-cni/blob/master/README.md) [NetworkAttachmentDefinition](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md) resources. For shared networks, these resources must be created in advance, while for tenant networks they are managed by the shoot reconciliation process.

In order to use any additional CNI plugins in a tenant network configuration, such as `bridge` or `firewall` in the above example, the plugin binaries must be present in the `/opt/cni/bin` directory of the provider cluster nodes. They can be installed manually by downloading a [containernetworking/plugins](https://github.com/containernetworking/plugins) release (not recommended except for testing a new configuration). Alternatively, they can be installed via a specially prepared daemon set that ensures the existence of the plugin binaries on each provider cluster node.

**Note:** Although it is possible to update the network configuration in `InfrastructureConfig`, any such changes will result in recreating all KubeVirt VMs, so that the new network configuration is properly taken into account. This will be done automatically by the MCM using rolling update.

## `ControlPlaneConfig`

The control plane configuration contains options for the KubeVirt-specific control plane components. Currently, the only component deployed by the KubeVirt extension is the [KubeVirt Cloud Controller Manager (CCM)](https://github.com/kubevirt/cloud-provider-kubevirt).

An example `ControlPlaneConfig` for the KubeVirt extension looks as follows:

```yaml
apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
kind: ControlPlaneConfig
cloudControllerManager:
  featureGates:
    CustomResourceValidation: true
```

The `cloudControllerManager.featureGates` contains a map of explicitly enabled or disabled feature gates. For production usage it's not recommend to use this field at all as you can enable alpha features or disable beta/stable features, potentially impacting the cluster stability. If you don't want to configure anything for the CCM, simply omit the key in the YAML specification.

## `WorkerConfig`

The KubeVirt extension supports specifying additional data volumes per machine in the worker pool. For each data volume, you must specify a name and a type. 

Below is an example `Shoot` resource snippet with root volume and data volumes: 

```yaml
spec:
  provider:
    workers:
    - name: cpu-worker
      ...
      volume:
        type: default
        size: 20Gi
      dataVolumes:
      - name: volume-1
        type: default
        size: 10Gi
```

**Note:** The additional data volumes will be attached as blank disks to the KubeVirt VMs. These disks must be formatted and mounted manually to the VM before they can be used. 

The KubeVirt extension does not currently support encryption for volumes. 

Additionally, it is possible to specify additional KubeVirt-specific options for configuring the worker pools. They can be specified in `.spec.provider.workers[].providerConfig` and are evaluated by the KubeVirt worker controller when it reconciles the shoot machines. 

An example `WorkerConfig` for the KubeVirt extension looks as follows:

```yaml
apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
kind: WorkerConfig
devices:
  # disks allow to customize disks attached to KubeVirt VM
  # check [link](https://kubevirt.io/user-guide/#/creation/disks-and-volumes?id=disks-and-volumes) for full specification and options
  disks:
  # name must match defined dataVolume name
  # to modify root volume the name must be equal to 'root-disk'
  - name: root-disk # modify root-disk
    # disk type, check [link](https://kubevirt.io/user-guide/#/creation/disks-and-volumes?id=disks) for more types
    disk:
      # bus indicates the type of disk device to emulate.
      bus: virtio
    # set disk device cache
    cache: writethrough
    # dedicatedIOThread indicates this disk should have an exclusive IO Thread
    dedicatedIOThread: true
  - name: volume-1 # modify dataVolume named volume-1
    disk: {}
  # whether to have random number generator from host
  rng: {}
  # whether or not to enable virtio multi-queue for block devices
  blockMultiQueue: true
  # if specified, virtual network interfaces configured with a virtio bus will also enable the vhost multiqueue feature
  networkInterfaceMultiQueue: true
cpu:
  # number of cores inside the VMI
  cores: 1
  # number of sockets inside the VMI
  sockets: 2
  # number of threads inside the VMI
  threads: 1
  # models specifies the CPU model of the VMI
  # list of available models https://github.com/libvirt/libvirt/tree/master/src/cpu_map.
  # and options https://libvirt.org/formatdomain.html#cpu-model-and-topology
  model: "host-model"
  # features specifies the CPU features list inside the VMI
  features:
  - "pcid"
  # dedicatedCPUPlacement requests the scheduler to place the VirtualMachineInstance on a node
  # with dedicated pCPUs and pin the vCPUs to it.
  dedicatedCpuPlacement: false
  # isolateEmulatorThread requests one more dedicated pCPU to be allocated for the VMI to place the emulator thread on it.
  isolateEmulatorThread: false
# memory configuration for KubeVirt VMs, allows to set 'hugepages' and 'guest' settings. 
# See https://kubevirt.io/api-reference/master/definitions.html#_v1_memory
memory:
  # hugepages requires appropriate feature gate to be enabled, take a look at the following links for more details:
  # * k8s - https://kubernetes.io/docs/tasks/manage-hugepages/scheduling-hugepages/
  # * okd - https://docs.okd.io/latest/scalability_and_performance/what-huge-pages-do-and-how-they-are-consumed-by-apps.html
  hugepages:
     pageSize: "2Mi"
  # guest allows to specifying the amount of memory which is visible inside the Guest OS. It must lie between requests and limits.
  # Defaults to the requested memory in the machineTypes.
  guest: "1Gi"
# overcommitGuestOverhead informs the scheduler to not take the guest-management overhead into account. Instead
# put the overhead only into the container's memory limit. This can lead to crashes if
# all memory is in use on a node. Defaults to false.
# For more details take a look at https://kubevirt.io/user-guide/#/usage/overcommit?id=overcommit-the-guest-overhead
overcommitGuestOverhead: true
# DNS policy for KubeVirt VMs. Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
# Defaults to 'ClusterFirst`.
# See https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
dnsPolicy: ClusterFirst
# DNS configuration for KubeVirt VMs, merged with the generated DNS configuration based on dnsPolicy.
# See https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
dnsConfig:
  nameservers:
  - 8.8.8.8
# Disable using pre-allocated data volumes. Defaults to 'false'.
disablePreAllocatedDataVolumes: true
# cpu allows to set the CPU topology of the VMI
# See https://kubevirt.io/api-reference/master/definitions.html#_v1_cpu
```

Currently, these KubeVirt-specific options may include:

* The CPU topology and memory configuration of the KubVirt VMs. For more information, see [CPU.v1](https://kubevirt.io/api-reference/master/definitions.html#_v1_cpu) and [Memory.v1](https://kubevirt.io/api-reference/master/definitions.html#_v1_memory). 
* The DNS policy and DNS configuration of the KubeVirt VMs. For more information, see [DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/).
* Whether to use *pre-allocated data volumes* with KubeVirt VMs. With pre-allocated data volumes (the default), a data volume is created in advance for each machine class, the OS image is imported into this volume only once, and actual KubeVirt VM data volumes are cloned from this data volume. Typically, this significantly speeds up the data volume creation process. You can disable this feature by setting the `disablePreAllocatedDataVolumes` option to `true`.

## Region and Zone Support

Nodes in the provider cluster may belong to provider-specific regions and zones, and Kubernetes would then use this information to spread pods across zones as described in [Running in multiple zones](https://kubernetes.io/docs/setup/best-practices/multiple-zones/). You may want to take advantage of these capabilities in the shoot cluster as well. 

To achieve this, the KubeVirt provider extension ensures that the region and zones specified in the `Shoot` resource are taken into account when creating the KubeVirt VMs used as shoot cluster nodes. 

Below is an example `Shoot` resource snippet with region and zones: 

```yaml
spec:  
  region: europe-west1
  provider:
    ...
    workers:
    - name: cpu-worker
      ...
      zones:
      - europe-west1-c
      - europe-west1-d
```

The shoot region and zones must correspond to the region and zones of the provider cluster. A KubeVirt VM designated for specific region and zone will only be scheduled on provider cluster nodes belonging to these region and zone. If there are no such nodes, or they have insufficient resources, the KubeVirt VM may remain in `Pending` state for a longer period and the shoot reconciliation may fail. Therefore, always make sure that the provider cluster contains nodes for all zones specified in the shoot. 

If multiple zones are specified for a worker pool, the KubeVirt VMs will be equally distributed over these zones in the specified order.   

If your provider cluster is not region and zone aware, or if it contains nodes that don't belong to any region or zone, you can use `default` as a region or zone name in the `Shoot` resource to target such nodes.

Note that the `region` and `zones` are mandatory fields in the `Shoot` resource, so you must specify either a concrete region / zone or `default`.

Once the KubeVirt VMs are scheduled on the correct provider cluster nodes, the KubeVirt Cloud Controller Manager (CCM) mentioned above will appropriately label the shoot worker nodes themselves with the appropriate [region and zone labels](https://kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/), by propagating the region and zone from the provider cluster nodes, so that Kubernetes multi-zone capabilities are also available in the shoot cluster.

## Example `Shoot` Manifest

Please find below an example `Shoot` manifest for one availability zone:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
metadata:
  name: johndoe-kubevirt
  namespace: garden-dev
spec:
  cloudProfileName: kubevirt
  secretBindingName: provider-cluster-kubeconfig
  region: europe-west1
  provider:
    type: kubevirt
#   infrastructureConfig:
#     apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
#     kind: InfrastructureConfig
#     networks:
#       tenantNetworks:
#       - name: network-1
#         config: "{...}"
#         default: true
#   controlPlaneConfig:
#     apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
#     kind: ControlPlaneConfig
#     cloudControllerManager:
#       featureGates:
#         CustomResourceValidation: true
    workers:
    - name: cpu-worker
      machine:
        type: standard-1
        image:
          name: ubuntu
          version: "18.04"
      minimum: 1
      maximum: 2
      volume:
        type: default
        size: 20Gi
#     dataVolumes:
#     - name: volume-1
#       type: default
#       size: 10Gi
#     providerConfig:
#       apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
#       kind: WorkerConfig
#       disablePreAllocatedDataVolumes: true
      zones:
      - europe-west1-c
  networking:
    type: calico
    pods: 100.96.0.0/11
    # Must match the IPAM subnet of the default tenant network, if present.
    # Otherwise, must be the same as the provider cluster pod network range.
    nodes: 10.225.128.0/17 # 10.100.0.0/16
    services: 100.64.0.0/13
  kubernetes:
    version: 1.17.8
  maintenance:
    autoUpdate:
      kubernetesVersion: true
      machineImageVersion: true
  addons:
    kubernetesDashboard:
      enabled: true
    nginxIngress:
      enabled: true
```

