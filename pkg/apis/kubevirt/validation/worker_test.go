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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("WorkerConfig validation", func() {
	var (
		nilPath *field.Path

		controlPlane *apiskubevirt.WorkerConfig
	)

	BeforeEach(func() {
		controlPlane = &apiskubevirt.WorkerConfig{}
	})

	Describe("#ValidateWorkerConfig", func() {
		It("should return no errors for a valid configuration", func() {
			Expect(ValidateWorkerConfig(controlPlane, nilPath)).To(BeEmpty())
		})
	})

	Describe("#ValidateWorkerConfigUpdate", func() {
		It("should return no errors for an unchanged config", func() {
			Expect(ValidateWorkerConfigUpdate(controlPlane, controlPlane, nilPath)).To(BeEmpty())
		})
	})
})
