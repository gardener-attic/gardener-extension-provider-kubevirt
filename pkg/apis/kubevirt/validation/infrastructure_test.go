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

package validation_test

import (
	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	. "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("InfrastructureConfig validation", func() {
	var (
		nilPath *field.Path

		infrastructureConfig *apiskubevirt.InfrastructureConfig
	)

	BeforeEach(func() {
		infrastructureConfig = &apiskubevirt.InfrastructureConfig{
			Networks: apiskubevirt.NetworksConfig{
				SharedNetworks: []apiskubevirt.NetworkAttachmentDefinitionReference{
					{
						Name:      "shared-1",
						Namespace: "default",
					},
					{
						Name:      "shared-2",
						Namespace: "default",
					},
				},
				TenantNetworks: []apiskubevirt.TenantNetwork{
					{
						Name:    "tenant-1",
						Config:  `{"type":"bridge"}`,
						Default: true,
					},
					{
						Name:   "tenant-2",
						Config: `{"type":"firewall"}`,
					},
				},
			},
		}
	})

	Describe("#ValidateInfrastructureConfig", func() {
		It("should return no errors for a valid configuration", func() {
			Expect(ValidateInfrastructureConfig(infrastructureConfig, nilPath)).To(BeEmpty())
		})

		It("should ensure that each shared network has a name", func() {
			infrastructureConfig.Networks.SharedNetworks[0].Name = ""

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("networks.sharedNetworks[0].name"),
			}))))
		})

		It("should ensure that there are duplicate shared networks", func() {
			infrastructureConfig.Networks.SharedNetworks[1].Name = "shared-1"

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeDuplicate),
				"Field": Equal("networks.sharedNetworks[1]"),
			}))))
		})

		It("should ensure that each tenant network has a name", func() {
			infrastructureConfig.Networks.TenantNetworks[0].Name = ""

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("networks.tenantNetworks[0].name"),
			}))))
		})

		It("should ensure that there are duplicate tenant network names", func() {
			infrastructureConfig.Networks.TenantNetworks[1].Name = "tenant-1"

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeDuplicate),
				"Field": Equal("networks.tenantNetworks[1]"),
			}))))
		})

		It("should ensure that each tenant network has a config", func() {
			infrastructureConfig.Networks.TenantNetworks[0].Config = ""

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("networks.tenantNetworks[0].config"),
			}))))
		})

		It("should ensure that each tenant network config is a valid JSON", func() {
			infrastructureConfig.Networks.TenantNetworks[0].Config = "abc"

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("networks.tenantNetworks[0].config"),
			}))))
		})

		It("should ensure that there is at most one default tenant network", func() {
			infrastructureConfig.Networks.TenantNetworks[1].Default = true

			errorList := ValidateInfrastructureConfig(infrastructureConfig, nilPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("networks.tenantNetworks[1].default"),
			}))))
		})
	})

	Describe("#ValidateInfrastructureConfigUpdate", func() {
		It("should return no errors for an unchanged config", func() {
			errorList := ValidateInfrastructureConfigUpdate(infrastructureConfig, infrastructureConfig, nilPath)
			Expect(errorList).To(BeEmpty())
		})
	})
})
