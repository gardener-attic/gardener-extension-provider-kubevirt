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
	"fmt"
	"strings"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"

	"github.com/gardener/gardener/pkg/apis/core"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateCloudProfileConfig validates a CloudProfileConfig object.
func ValidateCloudProfileConfig(profileSpec *core.CloudProfileSpec, profileConfig *apiskubevirt.CloudProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	machineImagesPath := field.NewPath("machineImages")
	if len(profileConfig.MachineImages) == 0 {
		allErrs = append(allErrs, field.Required(machineImagesPath, "must provide at least one machine image"))
	}
	machineImageVersions := map[string]sets.String{}
	for _, image := range profileSpec.MachineImages {
		versions := sets.String{}
		for _, version := range image.Versions {
			versions.Insert(version.Version)
		}
		machineImageVersions[image.Name] = versions
	}

	checkMachineImage := func(idxPath *field.Path, machineImage apiskubevirt.MachineImages) {
		definedVersions := sets.String{}
		if len(machineImage.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), "must provide a name"))
		}
		versions, ok := machineImageVersions[machineImage.Name]
		if !ok {
			allErrs = append(allErrs, field.Forbidden(idxPath.Child("name"), "machineImage with this name is not defined in cloud profile spec"))
		}

		if len(machineImage.Versions) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("versions"), fmt.Sprintf("must provide at least one version for machine image %q", machineImage.Name)))
		}
		for j, version := range machineImage.Versions {
			jdxPath := idxPath.Child("versions").Index(j)

			if len(version.Version) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("version"), "must provide a version"))
			} else {
				if definedVersions.Has(version.Version) {
					allErrs = append(allErrs, field.Duplicate(jdxPath.Child("version"), version.Version))
				}
				definedVersions.Insert(version.Version)
				if !versions.Has(version.Version) {
					allErrs = append(allErrs, field.Invalid(jdxPath.Child("version"), version.Version, "not defined as version in cloud profile spec"))
				}
			}
			if len(version.SourceURL) == 0 {
				allErrs = append(allErrs, field.Required(jdxPath.Child("sourceURL"), "must provide a source URL"))
			}
		}
		missing := versions.Difference(definedVersions)
		if missing.Len() > 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, strings.Join(missing.List(), ","), "missing versions"))
		}
	}
	for i, machineImage := range profileConfig.MachineImages {
		checkMachineImage(machineImagesPath.Index(i), machineImage)
	}

	return allErrs
}
