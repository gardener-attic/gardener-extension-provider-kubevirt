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

package kubevirt

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig is the infrastructure configuration resource.
type InfrastructureConfig struct {
	metav1.TypeMeta
	// Networks is the configuration of the infrastructure networks.
	Networks NetworksConfig
}

// NetworksConfig contains information about the configuration of the infrastructure networks.
type NetworksConfig struct {
	// SharedNetworks is a list of existing networks that can be shared between multiple clusters, e.g. storage networks.
	SharedNetworks []NetworkAttachmentDefinitionReference
	// TenantNetworks is a list of "tenant" networks that are only used by this cluster.
	TenantNetworks []TenantNetwork
}

// NetworkAttachmentDefinitionReference represents a NetworkAttachmentDefinition reference.
type NetworkAttachmentDefinitionReference struct {
	// Name is the name of the referenced NetworkAttachmentDefinition.
	Name string
	// Namespace is the namespace of the referenced NetworkAttachmentDefinition.
	Namespace string
}

// TenantNetwork represents a "tenant" network that is only used by a single cluster.
type TenantNetwork struct {
	// Name is the name of the tenant network.
	Name string
	// Config is the configuration of the tenant network.
	Config string
	// Default is whether the tenant network is the default or not.
	Default bool
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about the status of the infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta
	// Networks is the status of the infrastructure networks.
	Networks []NetworkStatus
}

// NetworkStatus contains information about the status of an infrastructure network.
type NetworkStatus struct {
	// Name is the name (in the format <name> or <namespace>/<name>) of the network.
	Name string
	// Default is whether the network is the default or not.
	Default bool
	// SHA is an SHA256 checksum of the network configuration.
	SHA string
}
