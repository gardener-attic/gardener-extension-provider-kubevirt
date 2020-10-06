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

package validation

import (
	"encoding/json"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(config *apiskubevirt.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	networksPath := fldPath.Child("networks")
	sharedNetworksPath := networksPath.Child("sharedNetworks")
	tenantNetworksPath := networksPath.Child("tenantNetworks")

	sharedNetworks := sets.String{}
	tenantNetworks := sets.String{}
	defaultTenantNetwork := ""

	checkSharedNetwork := func(path *field.Path, sharedNetwork apiskubevirt.NetworkAttachmentDefinitionReference) {
		// Ensure that the shared network has a name
		if len(sharedNetwork.Name) == 0 {
			allErrs = append(allErrs, field.Required(path.Child("name"), "must provide a name"))
		}

		// Ensure that there are no duplicate shared networks
		fullName := sharedNetwork.Namespace + "/" + sharedNetwork.Name
		if sharedNetworks.Has(fullName) {
			allErrs = append(allErrs, field.Duplicate(path, fullName))
		}
		sharedNetworks.Insert(fullName)
	}

	checkTenantNetwork := func(path *field.Path, tenantNetwork apiskubevirt.TenantNetwork) {
		// Ensure that the tenant network has a name
		if len(tenantNetwork.Name) == 0 {
			allErrs = append(allErrs, field.Required(path.Child("name"), "must provide a name"))
		}

		// Ensure that there are no duplicate tenant network names
		if tenantNetworks.Has(tenantNetwork.Name) {
			allErrs = append(allErrs, field.Duplicate(path, tenantNetwork.Name))
		}
		tenantNetworks.Insert(tenantNetwork.Name)

		// Ensure that the tenant network has a config
		if len(tenantNetwork.Config) == 0 {
			allErrs = append(allErrs, field.Required(path.Child("config"), "must provide a config"))
		} else {
			// Ensure that the tenant network config is a valid JSON
			config := make(map[string]interface{})
			if err := json.Unmarshal([]byte(tenantNetwork.Config), &config); err != nil {
				allErrs = append(allErrs, field.Invalid(path.Child("config"), tenantNetwork.Config, "must be a valid JSON"))
			}
		}

		// Ensure that there is at most one default tenant network
		if tenantNetwork.Default {
			if defaultTenantNetwork != "" {
				allErrs = append(allErrs, field.Invalid(path.Child("default"), tenantNetwork.Default, "there must be at most one default tenant network"))
			}
			defaultTenantNetwork = tenantNetwork.Name
		}
	}

	// Check shared and tenant networks
	for i, sharedNetwork := range config.Networks.SharedNetworks {
		checkSharedNetwork(sharedNetworksPath.Index(i), sharedNetwork)
	}
	for i, tenantNetwork := range config.Networks.TenantNetworks {
		checkTenantNetwork(tenantNetworksPath.Index(i), tenantNetwork)
	}

	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *apiskubevirt.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}
