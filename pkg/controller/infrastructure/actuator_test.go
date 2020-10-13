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

package infrastructure_test

import (
	"context"
	"encoding/json"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	kubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"
	. "github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"
	mockclient "github.com/gardener/gardener-extension-provider-kubevirt/pkg/mock/client"
	mockkubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/mock/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
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
	name      = "infrastructure"
	namespace = "shoot--dev--kubevirt"

	providerNamespace = "default"

	sharedNetworkName      = "net-conf"
	sharedNetworkNamespace = "default"
	sharedNetworkConfig    = `{}`

	tenantNetworkName     = "network-1"
	tenantNetworkFullName = namespace + "-" + tenantNetworkName
	tenantNetworkConfig   = `{"cniVersion": "0.4.0"}`

	oldTenantNetworkName     = "old-network"
	oldTenantNetworkFullName = namespace + "-" + oldTenantNetworkName
)

var _ = Describe("Actuator", func() {
	var (
		ctrl *gomock.Controller

		c              *mockclient.MockClient
		sw             *mockclient.MockStatusWriter
		networkManager *mockkubevirt.MockNetworkManager

		logger logr.Logger

		actuator infrastructure.Actuator

		labels = map[string]string{
			kubevirt.ClusterLabel: namespace,
		}

		newInfra = func(sharedNetworks []apiskubevirt.NetworkAttachmentDefinitionReference, tenantNetworks []apiskubevirt.TenantNetwork, networkStatuses []kubevirtv1alpha1.NetworkStatus) *extensionsv1alpha1.Infrastructure {
			var status extensionsv1alpha1.InfrastructureStatus
			if networkStatuses != nil {
				status = extensionsv1alpha1.InfrastructureStatus{
					DefaultStatus: extensionsv1alpha1.DefaultStatus{
						ProviderStatus: &runtime.RawExtension{
							Object: &kubevirtv1alpha1.InfrastructureStatus{
								TypeMeta: metav1.TypeMeta{
									APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
									Kind:       "InfrastructureStatus",
								},
								Networks: networkStatuses,
							},
						},
					},
				}
			}
			return &extensionsv1alpha1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: extensionsv1alpha1.InfrastructureSpec{
					SecretRef: corev1.SecretReference{
						Name:      v1beta1constants.SecretNameCloudProvider,
						Namespace: namespace,
					},
					DefaultSpec: extensionsv1alpha1.DefaultSpec{
						ProviderConfig: &runtime.RawExtension{
							Raw: encode(&apiskubevirt.InfrastructureConfig{
								Networks: apiskubevirt.NetworksConfig{
									SharedNetworks: sharedNetworks,
									TenantNetworks: tenantNetworks,
								},
							}),
						},
					},
				},
				Status: status,
			}
		}

		cluster = &extensionscontroller.Cluster{
			Shoot: &gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					Kubernetes: gardencorev1beta1.Kubernetes{
						Version: "1.13.4",
					},
				},
			},
		}

		sharedNetwork = apiskubevirt.NetworkAttachmentDefinitionReference{
			Name:      sharedNetworkName,
			Namespace: sharedNetworkNamespace,
		}
		tenantNetwork = apiskubevirt.TenantNetwork{
			Name:    tenantNetworkName,
			Config:  tenantNetworkConfig,
			Default: true,
		}

		sharedNetworkStatus = kubevirtv1alpha1.NetworkStatus{
			Name: sharedNetworkNamespace + "/" + sharedNetworkName,
			SHA:  utils.ComputeSHA256Hex([]byte(sharedNetworkConfig)),
		}
		tenantNetworkStatus = kubevirtv1alpha1.NetworkStatus{
			Name:    providerNamespace + "/" + namespace + "-" + tenantNetworkName,
			Default: true,
			SHA:     utils.ComputeSHA256Hex([]byte(tenantNetworkConfig)),
		}

		sharedNAD = &networkv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sharedNetworkName,
				Namespace: sharedNetworkNamespace,
			},
			Spec: networkv1.NetworkAttachmentDefinitionSpec{
				Config: sharedNetworkConfig,
			},
		}
		tenantNAD = &networkv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenantNetworkFullName,
				Namespace: providerNamespace,
			},
			Spec: networkv1.NetworkAttachmentDefinitionSpec{
				Config: tenantNetworkConfig,
			},
		}
		oldTenantNAD = &networkv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      oldTenantNetworkFullName,
				Namespace: providerNamespace,
			},
			Spec: networkv1.NetworkAttachmentDefinitionSpec{
				Config: tenantNetworkConfig,
			},
		}

		providerSecretData = map[string][]byte{
			kubevirt.KubeconfigSecretKey: []byte(kubeconfig),
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		sw = mockclient.NewMockStatusWriter(ctrl)
		c.EXPECT().Status().Return(sw).AnyTimes()
		networkManager = mockkubevirt.NewMockNetworkManager(ctrl)

		logger = log.Log.WithName("test")

		actuator = NewActuator(networkManager, logger)
		Expect(actuator.(inject.Client).InjectClient(c)).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Reconcile", func() {
		It("should create or update, delete, and get networks appropriately and update the infra status", func() {
			infra := newInfra([]apiskubevirt.NetworkAttachmentDefinitionReference{sharedNetwork}, []apiskubevirt.TenantNetwork{tenantNetwork}, nil)
			infraWithStatus := newInfra([]apiskubevirt.NetworkAttachmentDefinitionReference{sharedNetwork}, []apiskubevirt.TenantNetwork{tenantNetwork},
				[]kubevirtv1alpha1.NetworkStatus{tenantNetworkStatus, sharedNetworkStatus})

			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}, &corev1.Secret{}).
				DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
					secret.Data = providerSecretData
					return nil
				})

			networkManager.EXPECT().CreateOrUpdateNetworkAttachmentDefinition(context.TODO(), []byte(kubeconfig), tenantNetworkFullName, labels, tenantNetworkConfig).
				Return(tenantNAD, nil)
			networkManager.EXPECT().ListNetworkAttachmentDefinitions(context.TODO(), []byte(kubeconfig), labels).
				Return(&networkv1.NetworkAttachmentDefinitionList{
					Items: []networkv1.NetworkAttachmentDefinition{*tenantNAD, *oldTenantNAD},
				}, nil)
			networkManager.EXPECT().DeleteNetworkAttachmentDefinition(context.TODO(), []byte(kubeconfig), oldTenantNetworkFullName)
			networkManager.EXPECT().GetNetworkAttachmentDefinition(context.TODO(), []byte(kubeconfig), sharedNetworkName, sharedNetworkNamespace).
				Return(sharedNAD, nil)

			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, infra).Return(nil)
			sw.EXPECT().Update(context.TODO(), infraWithStatus).Return(nil)

			err := actuator.Reconcile(context.TODO(), infra, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#Delete", func() {
		It("should delete networks appropriately", func() {
			infra := newInfra([]apiskubevirt.NetworkAttachmentDefinitionReference{sharedNetwork}, []apiskubevirt.TenantNetwork{tenantNetwork}, nil)

			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}, &corev1.Secret{}).
				DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
					secret.Data = providerSecretData
					return nil
				})

			networkManager.EXPECT().DeleteNetworkAttachmentDefinition(context.TODO(), []byte(kubeconfig), tenantNetworkFullName)

			err := actuator.Delete(context.TODO(), infra, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

const kubeconfig = `apiVersion: v1
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

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
