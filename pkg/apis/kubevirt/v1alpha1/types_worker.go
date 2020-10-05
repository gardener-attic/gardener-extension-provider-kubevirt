// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerConfig contains configuration for VMs
type WorkerConfig struct {
	metav1.TypeMeta `json:",inline"`

	// CPU allows specifying the CPU topology of KubeVirt VM.
	// +optional
	CPU *kubevirtv1.CPU `json:"cpu,omitempty"`
	// Memory allows specifying the VirtualMachineInstance memory features like huge pages and guest memory settings.
	// Each feature might require appropriate FeatureGate enabled.
	// For hugepages take a look at:
	// k8s - https://kubernetes.io/docs/tasks/manage-hugepages/scheduling-hugepages/
	// okd - https://docs.okd.io/3.9/scaling_performance/managing_hugepages.html#huge-pages-prerequisites
	// +optional
	Memory *kubevirtv1.Memory `json:"memory,omitempty"`
	// Set DNS policy for the VM (the same as for the pod)
	// Defaults to "ClusterFirst".
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy
	// explicitly to 'ClusterFirstWithHostNet'.
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`
	// Specifies the DNS parameters of a VM.
	// Parameters specified here will be merged to the generated DNS
	// configuration based on DNSPolicy.
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty"`
	// DisablePreAllocatedDataVolumes disables using pre-allocated DataVolumes for VMs. Default is false.
	// +optional
	DisablePreAllocatedDataVolumes bool `json:"disablePreAllocatedDataVolumes,omitempty"`
	// OvercommitGuestOverhead informs the scheduler to not take the guest-management overhead into account. Instead
	// put the overhead only into the container's memory limit. This can lead to crashes if
	// all memory is in use on a node. Defaults to false.
	// +optional
	OvercommitGuestOverhead bool `json:"overcommitGuestOverhead,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerStatus contains information about created worker resources.
type WorkerStatus struct {
	metav1.TypeMeta `json:",inline"`

	// MachineImages is a list of machine images that have been used in this worker. Usually, the extension controller
	// gets the mapping from name/version to the provider-specific machine image data in its componentconfig. However, if
	// a version that is still in use gets removed from this componentconfig it cannot reconcile anymore existing `Worker`
	// resources that are still using this version. Hence, it stores the used versions in the provider status to ensure
	// reconciliation is possible.
	// +optional
	MachineImages []MachineImage `json:"machineImages,omitempty"`
}

// MachineImage is a mapping from logical names and versions to provider-specific machine image data.
type MachineImage struct {
	// Name is the logical name of the machine image.
	Name string `json:"name"`
	// Version is the logical version of the machine image.
	Version string `json:"version"`
	// SourceURL is the url of the machine image
	SourceURL string `json:"sourceUrl"`
}
