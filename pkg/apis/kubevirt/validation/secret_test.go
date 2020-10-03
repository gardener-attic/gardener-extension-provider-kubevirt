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
	. "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/validation"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Secret validation", func() {

	const validKubeconfig = `apiVersion: v1
kind: Config
current-context: provider
clusters:
- name: provider
  cluster:
    server: https://provider.example.com
contexts:
- name: provider
  context:
    cluster: provider
    user: admin
users:
- name: admin
  user:
    token: abc`

	DescribeTable("#ValidateCloudProviderSecret",
		func(data map[string][]byte, matcher gomegatypes.GomegaMatcher) {
			secret := &corev1.Secret{
				Data: data,
			}
			err := ValidateCloudProviderSecret(secret)

			Expect(err).To(matcher)
		},
		Entry("should return error when the kubeconfig field is missing",
			map[string][]byte{}, HaveOccurred()),
		Entry("should return error when the kubeconfig is not valid",
			map[string][]byte{kubevirt.Kubeconfig: []byte(`abc`)}, HaveOccurred()),
		Entry("should succeed when the kubeconfig is valid",
			map[string][]byte{kubevirt.Kubeconfig: []byte(validKubeconfig)}, BeNil()),
	)
})
