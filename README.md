# [Gardener Extension for KubeVirt provider](https://gardener.cloud)

[![CI Build status](https://concourse.ci.gardener.cloud/api/v1/teams/gardener/pipelines/gardener-extension-provider-kubevirt-master/jobs/master-head-update-job/badge)](https://concourse.ci.gardener.cloud/teams/gardener/pipelines/gardener-extension-provider-kubevirt-master/jobs/master-head-update-job)
[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/gardener-extension-provider-kubevirt)](https://goreportcard.com/report/github.com/gardener/gardener-extension-provider-kubevirt)

Project Gardener implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service.
Its main principle is to leverage Kubernetes concepts for all of its tasks.

Recently, most of the vendor specific logic has been developed [in-tree](https://github.com/gardener/gardener).
However, the project has grown to a size where it is very hard to extend, maintain, and test.
With [GEP-1](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md) we have proposed how the architecture can be changed in a way to support external controllers that contain their very own vendor specifics.
This way, we can keep Gardener core clean and independent.

This extension implements Gardener's extension contract for the [KubeVirt](https://kubevirt.io) provider. 
It includes KubeVirt-specific controllers for `Infrastructure`, `ControlPlane`, and `Worker` resources, as well as KubeVirt-specific control plane webhooks. 
Unlike other provider extensions, it does not include controllers for `BackupBucket` and `BackupEntry` resources, since KubeVirt as technology is not concerned with backup storage. 
Use the Gardener extension for your respective cloud provider to backup and restore your ETCD data. 
On OpenShift clusters, use [Gardener extension for OpenShift provider](https://github.com/gardener/gardener-extension-provider-openshift). 

For more information about Gardener integration with KubeVirt see [this gardener.cloud blog post](https://gardener.cloud/blog/2020-10/00/). 

An example for a `ControllerRegistration` resource that can be used to register the controllers of this extension with Gardener can be found [here](example/controller-registration.yaml).

Please find more information regarding the extensibility concepts and a detailed proposal [here](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md).

## Supported Kubernetes versions

This extension supports the following Kubernetes versions:

| Version         | Support     | Conformance test results |
| --------------- | ----------- | ------------------------ |
| Kubernetes 1.19 | not tested  | N/A |
| Kubernetes 1.18 | 1.18.0+     | N/A |
| Kubernetes 1.17 | 1.17.0+     | N/A |
| Kubernetes 1.16 | not tested  | N/A |
| Kubernetes 1.15 | not tested  | N/A |

Please take a look [here](https://github.com/gardener/gardener/blob/master/docs/usage/supported_k8s_versions.md) to see which versions are supported by Gardener in general.

----

## How to start using or developing this extension locally

You can run the extension locally on your machine by executing `make start`.

Static code checks and tests can be executed by running `make verify`. We are using Go modules for Golang package dependency management and [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) for testing.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/gardener-extension-provider-kubevirt/issues) or join our [Slack channel #gardener](https://kubernetes.slack.com/messages/gardener) (please invite yourself to the Kubernetes workspace [here](http://slack.k8s.io)).

## Learn more!

Please find further resources about out project here:

* [Our landing page gardener.cloud](https://gardener.cloud/)
* ["Gardener, the Kubernetes Botanist" blog on kubernetes.io](https://kubernetes.io/blog/2018/05/17/gardener/)
* ["Gardener Project Update" blog on kubernetes.io](https://kubernetes.io/blog/2019/12/02/gardener-project-update/)
* [GEP-1 (Gardener Enhancement Proposal) on extensibility](https://github.com/gardener/gardener/blob/master/docs/proposals/01-extensibility.md)
* [GEP-4 (New `core.gardener.cloud/v1alpha1` API)](https://github.com/gardener/gardener/blob/master/docs/proposals/04-new-core-gardener-cloud-apis.md)
* [Extensibility API documentation](https://github.com/gardener/gardener/tree/master/docs/extensions)
* [Gardener Extensions Golang library](https://godoc.org/github.com/gardener/gardener/extensions/pkg)
* [Gardener API Reference](https://gardener.cloud/api-reference/)
