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

package controlplane

import (
	"context"
	"encoding/json"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	namespace = "test"
)

var _ = Describe("ValuesProvider", func() {
	var (
		ctrl   *gomock.Controller
		ctx    = context.TODO()
		logger = log.Log.WithName("test")

		c  *mockclient.MockClient
		vp genericactuator.ValuesProvider

		scheme = runtime.NewScheme()
		_      = apiskubevirt.AddToScheme(scheme)

		infrastructureStatus = &apiskubevirt.InfrastructureStatus{}

		cp = &extensionsv1alpha1.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "control-plane",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.ControlPlaneSpec{
				Region: "local",
				SecretRef: corev1.SecretReference{
					Name:      v1beta1constants.SecretNameCloudProvider,
					Namespace: namespace,
				},
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					ProviderConfig: &runtime.RawExtension{
						Raw: encode(&apiskubevirt.ControlPlaneConfig{
							CloudControllerManager: &apiskubevirt.CloudControllerManagerConfig{
								FeatureGates: map[string]bool{
									"CustomResourceValidation": true,
								},
							},
						}),
					},
				},
				InfrastructureProviderStatus: &runtime.RawExtension{
					Raw: encode(infrastructureStatus),
				},
			},
		}

		cidr    = "10.250.0.0/19"
		cluster = &extensionscontroller.Cluster{
			Shoot: &gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					Networking: gardencorev1beta1.Networking{
						Pods: &cidr,
					},
					Kubernetes: gardencorev1beta1.Kubernetes{
						Version: "1.13.4",
					},
				},
			},
		}

		cpSecretKey = client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}
		cpSecret    = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v1beta1constants.SecretNameCloudProvider,
				Namespace: namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"kubeconfig": []byte(`kubeconfig`),
			},
		}

		checksums = map[string]string{
			kubevirt.CloudControllerManagerName:             "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
			kubevirt.CloudControllerManagerName + "-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
			kubevirt.CloudProviderConfigName:                "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
		}

		enabledTrue = map[string]interface{}{"enabled": true}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		vp = NewValuesProvider(logger)

		err := vp.(inject.Scheme).InjectScheme(scheme)
		Expect(err).NotTo(HaveOccurred())
		err = vp.(inject.Client).InjectClient(c)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#GetConfigChartValues", func() {
		It("should return correct config chart values", func() {
			c.EXPECT().Get(ctx, cpSecretKey, &corev1.Secret{}).DoAndReturn(clientGet(cpSecret))

			values, err := vp.GetConfigChartValues(ctx, cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(map[string]interface{}{
				"kubeconfig": "kubeconfig",
			}))
		})
	})

	Describe("#GetControlPlaneChartValues", func() {
		ccmChartValues := utils.MergeMaps(enabledTrue, map[string]interface{}{
			"replicas":          1,
			"clusterName":       namespace,
			"kubernetesVersion": "1.17.5",
			"podNetwork":        cidr,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-cloud-controller-manager":        "3d791b164a808638da9a8df03924be2a41e34cd664e42231c00fe369e3588272",
				"checksum/secret-cloud-controller-manager-server": "6dff2a2e6f14444b66d8e4a351c049f7e89ee24ba3eaab95dbec40ba6bdebb52",
				"checksum/secret-cloud-provider-config":           "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432",
			},
			"podLabels": map[string]interface{}{
				"maintenance.gardener.cloud/restart": "true",
			},
			"featureGates": map[string]bool{
				"CustomResourceValidation": true,
			},
		})

		It("should return correct control plane chart values", func() {
			values, err := vp.GetControlPlaneChartValues(ctx, cp, cluster, checksums, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(map[string]interface{}{
				kubevirt.CloudControllerManagerName: utils.MergeMaps(ccmChartValues, map[string]interface{}{
					"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
				}),
			}))
		})
	})

	Describe("#GetControlPlaneShootChartValues", func() {
		It("should return correct control plane shoot chart values", func() {
			values, err := vp.GetControlPlaneShootChartValues(ctx, cp, cluster, checksums)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(map[string]interface{}{
				kubevirt.CloudControllerManagerName: enabledTrue,
			}))
		})
	})

	Describe("#GetStorageClassesChartValues()", func() {
		It("should return empty storage class chart values", func() {
			values, err := vp.GetStorageClassesChartValues(ctx, cp, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(map[string]interface{}{}))
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}

func clientGet(result runtime.Object) interface{} {
	return func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
		switch obj.(type) {
		case *corev1.Secret:
			*obj.(*corev1.Secret) = *result.(*corev1.Secret)
		}
		return nil
	}
}
