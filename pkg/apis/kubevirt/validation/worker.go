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

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateWorkerConfig validates a WorkerConfig object.
func ValidateWorkerConfig(config *apiskubevirt.WorkerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if config.DNSPolicy != "" {
		dnsPolicyPath := fldPath.Child("dnsPolicy")
		dnsConfigPath := fldPath.Child("dnsConfig")

		switch config.DNSPolicy {
		case corev1.DNSDefault, corev1.DNSClusterFirstWithHostNet, corev1.DNSClusterFirst, corev1.DNSNone:
			break
		default:
			allErrs = append(allErrs, field.Invalid(dnsPolicyPath, config.DNSPolicy, "invalid dns policy"))
		}

		if config.DNSPolicy == corev1.DNSNone {
			if config.DNSConfig != nil {
				if len(config.DNSConfig.Nameservers) == 0 {
					allErrs = append(allErrs, field.Required(dnsConfigPath.Child("nameservers"),
						fmt.Sprintf("cannot be empty when dns policy is %s", corev1.DNSNone)))
				}
			} else {
				allErrs = append(allErrs, field.Required(dnsConfigPath,
					fmt.Sprintf("cannot be empty when dns policy is %s", corev1.DNSNone)))
			}
		}
	}

	return allErrs
}

// ValidateWorkerConfigUpdate validates a WorkerConfig object.
func ValidateWorkerConfigUpdate(oldConfig, newConfig *apiskubevirt.WorkerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}
