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

	"github.com/gardener/gardener/pkg/apis/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("CloudProfileConfig validation", func() {
	Describe("#ValidateCloudProfileConfig", func() {
		var cloudProfileConfig *apiskubevirt.CloudProfileConfig
		var cloudProfileSpec *core.CloudProfileSpec

		BeforeEach(func() {
			cloudProfileConfig = &apiskubevirt.CloudProfileConfig{
				MachineImages: []apiskubevirt.MachineImages{
					{
						Name: "ubuntu",
						Versions: []apiskubevirt.MachineImageVersion{
							{
								Version:   "16.04",
								SourceURL: "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img",
							},
							{
								Version:   "18.04",
								SourceURL: "https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img",
							},
						},
					},
				},
			}
			cloudProfileSpec = &core.CloudProfileSpec{
				MachineImages: []core.MachineImage{
					{
						Name: "ubuntu",
						Versions: []core.MachineImageVersion{
							{
								ExpirableVersion: core.ExpirableVersion{
									Version: "16.04",
								},
							},
							{
								ExpirableVersion: core.ExpirableVersion{
									Version: "18.04",
								},
							},
						},
					},
				},
			}
		})

		Context("machine image validation", func() {
			It("should validate valid machine image version configuration", func() {
				errorList := ValidateCloudProfileConfig(cloudProfileSpec, cloudProfileConfig)
				Expect(errorList).To(ConsistOf())
			})

			It("should validate valid machine image version configuration", func() {
				errorList := ValidateCloudProfileConfig(cloudProfileSpec, cloudProfileConfig)
				Expect(errorList).To(ConsistOf())
			})

			It("should enforce that at least one machine image has been defined", func() {
				cloudProfileConfig.MachineImages = []apiskubevirt.MachineImages{}

				errorList := ValidateCloudProfileConfig(cloudProfileSpec, cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages"),
				}))))
			})

			It("should forbid unsupported machine image configuration", func() {
				cloudProfileConfig.MachineImages = []apiskubevirt.MachineImages{{}}

				errorList := ValidateCloudProfileConfig(cloudProfileSpec, cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].name"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("machineImages[0].name"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions"),
				}))))
			})

			It("should forbid unsupported machine image version configuration", func() {
				cloudProfileConfig.MachineImages = []apiskubevirt.MachineImages{
					{
						Name:     "abc",
						Versions: []apiskubevirt.MachineImageVersion{{}},
					},
				}

				errorList := ValidateCloudProfileConfig(cloudProfileSpec, cloudProfileConfig)

				Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("machineImages[0].name"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions[0].version"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("machineImages[0].versions[0].sourceURL"),
				}))))
			})
		})
	})
})
