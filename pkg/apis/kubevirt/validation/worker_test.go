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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
				err := ValidateWorkerConfig(config, nilPath)

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

	})
})
