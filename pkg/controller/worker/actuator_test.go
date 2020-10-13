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

package worker_test

import (
	"context"

	. "github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/worker"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"
	mockclient "github.com/gardener/gardener-extension-provider-kubevirt/pkg/mock/client"
	mockkubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/mock/kubevirt"
	mockworker "github.com/gardener/gardener-extension-provider-kubevirt/pkg/mock/worker"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const (
	name      = "worker"
	namespace = "shoot--dev--kubevirt"

	providerNamespace = "default"

	workerPoolName = "worker-1"

	machineClassName    = namespace + "-" + workerPoolName + "-z1-aaaaa"
	oldMachineClassName = namespace + "-" + workerPoolName + "-z1-bbbbb"
)

var _ = Describe("Actuator", func() {
	var (
		ctrl *gomock.Controller

		c                 *mockclient.MockClient
		dataVolumeManager *mockkubevirt.MockDataVolumeManager
		workerActuator    *mockworker.MockActuator

		logger logr.Logger

		actuator worker.Actuator

		labels = map[string]string{
			kubevirt.ClusterLabel: namespace,
		}

		w = &extensionsv1alpha1.Worker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.WorkerSpec{
				SecretRef: corev1.SecretReference{
					Name:      v1beta1constants.SecretNameCloudProvider,
					Namespace: namespace,
				},
				Pools: []extensionsv1alpha1.WorkerPool{
					{
						Name:  workerPoolName,
						Zones: []string{"zone-1"},
					},
				},
			},
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

		machineClass = &machinev1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineClassName,
				Namespace: namespace,
			},
		}

		dataVolume = &cdicorev1alpha1.DataVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineClassName,
				Namespace: providerNamespace,
			},
		}
		oldDataVolume = &cdicorev1alpha1.DataVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      oldMachineClassName,
				Namespace: providerNamespace,
			},
		}

		providerSecretData = map[string][]byte{
			kubevirt.KubeconfigSecretKey: []byte(kubeconfig),
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		dataVolumeManager = mockkubevirt.NewMockDataVolumeManager(ctrl)
		workerActuator = mockworker.NewMockActuator(ctrl)

		logger = log.Log.WithName("test")

		actuator = NewActuator(workerActuator, dataVolumeManager, logger)
		Expect(actuator.(inject.Client).InjectClient(c)).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Reconcile", func() {
		It("should delete orphaned data volumes appropriately", func() {
			workerActuator.EXPECT().Reconcile(context.TODO(), w, cluster)

			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}, &corev1.Secret{}).
				DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
					secret.Data = providerSecretData
					return nil
				})
			c.EXPECT().List(context.TODO(), &machinev1alpha1.MachineClassList{}, client.InNamespace(namespace)).
				DoAndReturn(func(_ context.Context, machineClasses *machinev1alpha1.MachineClassList, _ ...client.ListOption) error {
					machineClasses.Items = []machinev1alpha1.MachineClass{*machineClass}
					return nil
				})

			dataVolumeManager.EXPECT().ListDataVolumes(context.TODO(), []byte(kubeconfig), labels).
				Return(&cdicorev1alpha1.DataVolumeList{
					Items: []cdicorev1alpha1.DataVolume{*dataVolume, *oldDataVolume},
				}, nil)
			dataVolumeManager.EXPECT().DeleteDataVolume(context.TODO(), []byte(kubeconfig), oldMachineClassName)

			err := actuator.Reconcile(context.TODO(), w, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#Delete", func() {
		It("should delete orphaned data volumes appropriately", func() {
			workerActuator.EXPECT().Delete(context.TODO(), w, cluster)

			c.EXPECT().Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: v1beta1constants.SecretNameCloudProvider}, &corev1.Secret{}).
				DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
					secret.Data = providerSecretData
					return nil
				})
			c.EXPECT().List(context.TODO(), &machinev1alpha1.MachineClassList{}, client.InNamespace(namespace)).
				DoAndReturn(func(_ context.Context, machineClasses *machinev1alpha1.MachineClassList, _ ...client.ListOption) error {
					machineClasses.Items = []machinev1alpha1.MachineClass{}
					return nil
				})

			dataVolumeManager.EXPECT().ListDataVolumes(context.TODO(), []byte(kubeconfig), labels).
				Return(&cdicorev1alpha1.DataVolumeList{
					Items: []cdicorev1alpha1.DataVolume{*dataVolume},
				}, nil)
			dataVolumeManager.EXPECT().DeleteDataVolume(context.TODO(), []byte(kubeconfig), machineClassName)

			err := actuator.Delete(context.TODO(), w, cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
