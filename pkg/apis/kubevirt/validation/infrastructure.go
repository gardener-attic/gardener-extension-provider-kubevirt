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
	api "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(infra *api.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// TODO Validate networks
	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *api.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// TODO Ensures that networks are immutable

	return allErrs
}

// ValidateInfrastructureConfigAgainstCloudProfile validates the given InfrastructureConfig against constraints in the given CloudProfile.
func ValidateInfrastructureConfigAgainstCloudProfile(infra *api.InfrastructureConfig, shootRegion string, cloudProfileConfig *api.CloudProfileConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

// HasRelevantInfrastructureConfigUpdates returns true if given InfrastructureConfig has relevant changes
func HasRelevantInfrastructureConfigUpdates(oldInfra *api.InfrastructureConfig, newInfra *api.InfrastructureConfig) bool {
	return false
}
