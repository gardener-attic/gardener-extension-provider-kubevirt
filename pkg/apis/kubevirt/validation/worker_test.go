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
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	gardenercore "github.com/gardener/gardener/pkg/apis/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

var _ = Describe("WorkerConfig validation", func() {
	var (
		nilPath *field.Path

		config *apiskubevirt.WorkerConfig
	)

	BeforeEach(func() {
		config = &apiskubevirt.WorkerConfig{}
	})

	Describe("#ValidateWorkerConfig", func() {

		DescribeTable("#ValidateDNS",
			func(dnsPolicy corev1.DNSPolicy, dnsConfig *corev1.PodDNSConfig, matcher gomegatypes.GomegaMatcher) {
				config.DNSPolicy = dnsPolicy
				config.DNSConfig = dnsConfig
				err := ValidateWorkerConfig(config, nil, nilPath)

				Expect(err).To(matcher)
			},
			Entry("should return error when invalid DNS is set",
				corev1.DNSPolicy("invalid-policy"), nil, ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("dnsPolicy"),
				}))),
			),
			Entry("should return error when dnsConfig is empty with 'None' dnsPolicy",
				corev1.DNSNone, nil, ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("dnsConfig"),
				}))),
			),
			Entry("should return error when dnsConfig.nameservers is empty with 'None' dnsPolicy",
				corev1.DNSNone, &corev1.PodDNSConfig{}, ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("dnsConfig.nameservers"),
				}))),
			),
			Entry("should not return error when dnsConfig.nameservers is set with 'None' dnsPolicy",
				corev1.DNSNone, &corev1.PodDNSConfig{Nameservers: []string{"8.8.8.8"}}, Equal(field.ErrorList{}),
			),
			Entry("should not return error with appropriate dnsPolicy",
				corev1.DNSDefault, nil, Equal(field.ErrorList{}),
			),
			Entry("should not return error with appropriate dnsPolicy and dnsConfig",
				corev1.DNSDefault, &corev1.PodDNSConfig{Nameservers: []string{"8.8.8.8"}}, Equal(field.ErrorList{}),
			),
		)

		bootOrder0 := uint(0)
		DescribeTable("#ValidateDisksAndVolumes",
			func(disks []kubevirtv1.Disk, dataVolumes []gardenercore.DataVolume, matcher gomegatypes.GomegaMatcher) {
				config.Devices = &apiskubevirt.Devices{
					Disks: disks,
				}
				err := ValidateWorkerConfig(config, dataVolumes, nilPath)
				Expect(err).To(matcher)
			},
			Entry("should not return error with appropriate disks and volumes match",
				[]kubevirtv1.Disk{
					{
						Name: "disk-1",
					},
					{
						Name: "disk-2",
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-1",
					},
					{
						Name: "disk-2",
					},
				},
				Equal(field.ErrorList{}),
			),
			Entry("should return error with empty disk name",
				[]kubevirtv1.Disk{
					{
						Name: "",
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-1",
					},
				},
				ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("devices.disks[0].name"),
				}))),
			),
			Entry("should return error with disks and volumes that do not match",
				[]kubevirtv1.Disk{
					{
						Name: "disk-1a",
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-1b",
					},
				},
				ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("devices.disks[0].name"),
				}))),
			),
			Entry("should return error with number of disks bigger than volumes",
				[]kubevirtv1.Disk{
					{
						Name: kubevirt.RootDiskName,
					},
					{
						Name: "disk-1",
					},
					{
						Name: "disk-2",
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-2",
					},
				},
				ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("devices.disks"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("devices.disks[1].name"),
					})),
				),
			),
			Entry("should return error with duplicated disks names",
				[]kubevirtv1.Disk{
					{
						Name: kubevirt.RootDiskName,
					},
					{
						Name: kubevirt.RootDiskName,
					},
					{
						Name: "disk-1",
					},
					{
						Name: "disk-1",
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-1",
					},
					{
						Name: "disk-2",
					},
					{
						Name: "disk-3",
					},
				},
				ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("devices.disks[1].name"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("devices.disks[3].name"),
					})),
				),
			),
			Entry("should return error when boot order is set for disk",
				[]kubevirtv1.Disk{
					{
						Name: "disk-1",
					},
					{
						Name:      "disk-2",
						BootOrder: &bootOrder0,
					},
				},
				[]gardenercore.DataVolume{
					{
						Name: "disk-1",
					},
					{
						Name: "disk-2",
					},
				},
				ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeForbidden),
						"Field": Equal("devices.disks[1].bootOrder"),
					})),
				),
			),
		)

	})
})
