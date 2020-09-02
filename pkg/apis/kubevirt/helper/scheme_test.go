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

package helper_test

import (
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/install"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Helper (decode)", func() {
	var (
		s                  *runtime.Scheme
		ctx                common.ClientContext
		cluster            *controller.Cluster
		cloudProfileConfig *kubevirt.CloudProfileConfig
	)

	BeforeEach(func() {
		s = scheme.Scheme
		install.Install(s)
		ctx = common.ClientContext{}
		err := ctx.InjectScheme(s)
		if err != nil {
			panic(err)
		}

		cluster = &controller.Cluster{
			Shoot: &gardencorev1beta1.Shoot{
				ObjectMeta: v1.ObjectMeta{Name: "test"},
			},
			CloudProfile: &gardencorev1beta1.CloudProfile{
				Spec: gardencorev1beta1.CloudProfileSpec{
					MachineImages: []gardencorev1beta1.MachineImage{
						{
							Name: "ubuntu",
							Versions: []gardencorev1beta1.ExpirableVersion{
								{
									Version: "16.04",
								},
							},
						},
					},
					ProviderConfig: &runtime.RawExtension{Raw: []byte(`
apiVersion: kubevirt.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
machineImages:
- name: "ubuntu"
  versions:
  - version: "16.04"
    sourceURL: "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img"
`)},
				},
			},
		}

		cloudProfileConfig = &kubevirt.CloudProfileConfig{
			MachineImages: []kubevirt.MachineImages{
				{
					Name: "ubuntu",
					Versions: []kubevirt.MachineImageVersion{
						{
							Version:   "16.04",
							SourceURL: "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img",
						},
					},
				},
			},
		}
	})

	Describe("#GetCloudProfileConfig", func() {
		It("should decode the CloudProfileConfig", func() {
			result, err := helper.GetCloudProfileConfig(cluster.CloudProfile)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(cloudProfileConfig))
		})
	})
})
