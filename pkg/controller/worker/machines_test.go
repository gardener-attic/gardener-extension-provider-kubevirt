/*
 * Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	kubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"
	. "github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/worker"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener/pkg/mock/gardener/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Machines", func() {
	var (
		ctrl           *gomock.Controller
		c              *mockclient.MockClient
		chartApplier   *mockkubernetes.MockChartApplier
		providerClient *mockclient.MockClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		c = mockclient.NewMockClient(ctrl)
		providerClient = mockclient.NewMockClient(ctrl)
		chartApplier = mockkubernetes.NewMockChartApplier(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	mockProviderClient := func(kubeconfig []byte) (client.Client, string, error) {
		return providerClient, "", nil
	}

	Context("with workerDelegate", func() {
		workerDelegate, _ := NewWorkerDelegate(common.NewClientContext(nil, nil, nil), nil, "", nil, nil, nil)

		Describe("#MachineClassKind", func() {
			It("should return the correct kind of the machine class", func() {
				Expect(workerDelegate.MachineClassKind()).To(Equal("MachineClass"))
			})
		})

		Describe("#MachineClassList", func() {
			It("should return the correct type for the machine class list", func() {
				Expect(workerDelegate.MachineClassList()).To(Equal(&machinev1alpha1.MachineClassList{}))
			})
		})

		Describe("#GenerateMachineDeployments, #DeployMachineClasses", func() {
			var (
				scheme                           *runtime.Scheme
				decoder                          runtime.Decoder
				cluster                          *extensionscontroller.Cluster
				workerPoolHash1, workerPoolHash2 string
				dataVolumeManager                kubevirt.DataVolumeManager
				err                              error
			)

			namespace := "shoot--dev--kubevirt-1"
			machineType1, machineType2 := "local-1", "local-2"
			namePool1, namePool2 := "pool-1", "pool-2"
			minPool1, minPool2 := int32(5), int32(3)
			maxPool1, maxPool2 := int32(7), int32(6)
			maxSurgePool1, maxSurgePool2 := intstr.FromInt(3), intstr.FromInt(5)
			maxUnavailablePool1, maxUnavailablePool2 := intstr.FromInt(2), intstr.FromInt(4)
			machineImageName := "ubuntu"
			machineImageVersion := "16.04"
			userData := []byte("user-data")
			shootVersion := "1.2.3"
			cloudProfileName := "test-profile"
			ubuntuSourceURL := "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img"
			sshPublicKey := []byte("ssh-rsa AAAAB3...")
			machineConfiguration := &machinev1alpha1.MachineConfiguration{}
			networkName := "default/net-conf"
			networkSHA := "abc"
			dnsNameserver := "8.8.8.8"

			images := []kubevirtv1alpha1.MachineImages{
				{
					Name: machineImageName,
					Versions: []kubevirtv1alpha1.MachineImageVersion{
						{
							Version:   machineImageVersion,
							SourceURL: ubuntuSourceURL,
						},
					},
				},
			}

			w := &extensionsv1alpha1.Worker{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
				},
				Spec: extensionsv1alpha1.WorkerSpec{
					SecretRef: corev1.SecretReference{
						Name:      "kubevirt-provider-credentials",
						Namespace: namespace,
					},
					Region: "local",
					InfrastructureProviderStatus: &runtime.RawExtension{
						Raw: encode(&kubevirtv1alpha1.InfrastructureStatus{
							TypeMeta: metav1.TypeMeta{
								APIVersion: "kubevirt.provider.extensions.gardener.cloud/v1alpha1",
								Kind:       "InfrastructureStatus",
							},
							Networks: []kubevirtv1alpha1.NetworkStatus{
								{
									Name:    networkName,
									Default: true,
									SHA:     networkSHA,
								},
							},
						}),
					},
					Pools: []extensionsv1alpha1.WorkerPool{
						{
							Name:           namePool1,
							Minimum:        minPool1,
							Maximum:        maxPool1,
							MaxSurge:       maxSurgePool1,
							MaxUnavailable: maxUnavailablePool1,
							MachineType:    machineType1,
							MachineImage: extensionsv1alpha1.MachineImage{
								Name:    machineImageName,
								Version: machineImageVersion,
							},
							UserData: userData,
							Zones:    []string{"local-1", "local-2"},
							ProviderConfig: &runtime.RawExtension{
								Raw: encode(&kubevirtv1alpha1.WorkerConfig{
									TypeMeta: metav1.TypeMeta{
										APIVersion: "kubevirt.provider.extensions.gardener.cloud/v1alpha1",
										Kind:       "WorkerConfig",
									},
									OvercommitGuestOverhead: true,
									DNSPolicy:               corev1.DNSDefault,
									DNSConfig: &corev1.PodDNSConfig{
										Nameservers: []string{dnsNameserver},
									},
									DisablePreAllocatedDataVolumes: true,
									Memory: &kubevirtv1.Memory{
										Hugepages: &kubevirtv1.Hugepages{
											PageSize: "2Mi",
										},
									},
									CPU: &kubevirtv1.CPU{
										Cores:                 uint32(1),
										Sockets:               uint32(2),
										Threads:               uint32(1),
										DedicatedCPUPlacement: true,
									},
								}),
							},
						},
						{
							Name:           namePool2,
							Minimum:        minPool2,
							Maximum:        maxPool2,
							MaxSurge:       maxSurgePool2,
							MaxUnavailable: maxUnavailablePool2,
							MachineType:    machineType2,
							MachineImage: extensionsv1alpha1.MachineImage{
								Name:    machineImageName,
								Version: machineImageVersion,
							},
							UserData: userData,
							Zones:    []string{"local-3"},
						},
					},
					SSHPublicKey: sshPublicKey,
				},
			}

			It("should return a data volume manager", func() {
				dataVolumeManager, err = kubevirt.NewDefaultDataVolumeManager(kubevirt.ClientFactoryFunc(mockProviderClient))
				Expect(err).NotTo(HaveOccurred())
			})

			BeforeEach(func() {
				scheme = runtime.NewScheme()
				_ = apiskubevirt.AddToScheme(scheme)
				_ = kubevirtv1alpha1.AddToScheme(scheme)
				decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()

				cluster = createCluster(cloudProfileName, shootVersion, images)

				workerPoolHash1, _ = worker.WorkerPoolHash(w.Spec.Pools[0], cluster, networkName, "true", networkSHA)
				workerPoolHash2, _ = worker.WorkerPoolHash(w.Spec.Pools[1], cluster, networkName, "true", networkSHA)
				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster, dataVolumeManager)
			})

			It("should return the expected machine deployments", func() {
				generateKubeVirtSecret(c)

				machineDeploymentName1 := fmt.Sprintf("%s-%s-z%d", namespace, namePool1, 1)
				machineDeploymentName2 := fmt.Sprintf("%s-%s-z%d", namespace, namePool1, 2)
				machineDeploymentName3 := fmt.Sprintf("%s-%s-z%d", namespace, namePool2, 1)

				machineClassName1 := fmt.Sprintf("%s-%s", machineDeploymentName1, workerPoolHash1)
				machineClassName2 := fmt.Sprintf("%s-%s", machineDeploymentName2, workerPoolHash1)
				machineClassName3 := fmt.Sprintf("%s-%s", machineDeploymentName3, workerPoolHash2)

				machineClassTemplate := map[string]interface{}{
					"storageClassName": "standard",
					"sourceURL":        ubuntuSourceURL,
					"sshKeys": []string{
						string(sshPublicKey),
					},
					"networks": []kubevirtv1alpha1.NetworkStatus{
						{
							Name:    networkName,
							Default: true,
							SHA:     networkSHA,
						},
					},
					"region": "local",
					"secret": map[string]interface{}{
						"cloudConfig": "user-data",
						"kubeconfig":  kubeconfig,
					},
					"tags": map[string]string{
						"mcm.gardener.cloud/cluster":      namespace,
						"mcm.gardener.cloud/role":         "node",
						"mcm.gardener.cloud/machineclass": machineClassName1,
					},
				}

				machineClass1 := generateMachineClass(
					machineClassTemplate,
					machineClassName1,
					"8Gi",
					"local-1",
					&kubevirtv1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4096Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
						OvercommitGuestOverhead: true,
					},
					map[string]string{
						"mcm.gardener.cloud/cluster":      namespace,
						"mcm.gardener.cloud/role":         "node",
						"mcm.gardener.cloud/machineclass": machineClassName1,
					},
					corev1.DNSDefault,
					&corev1.PodDNSConfig{
						Nameservers: []string{dnsNameserver},
					},
					true,
					&kubevirtv1.Memory{
						Hugepages: &kubevirtv1.Hugepages{
							PageSize: "2Mi",
						},
					},
					&kubevirtv1.CPU{
						Cores:                 uint32(1),
						Sockets:               uint32(2),
						Threads:               uint32(1),
						DedicatedCPUPlacement: true,
					},
				)

				machineClass2 := generateMachineClass(
					machineClassTemplate,
					machineClassName2,
					"8Gi",
					"local-2",
					&kubevirtv1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("4096Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
						OvercommitGuestOverhead: true,
					},
					map[string]string{
						"mcm.gardener.cloud/cluster":      namespace,
						"mcm.gardener.cloud/role":         "node",
						"mcm.gardener.cloud/machineclass": machineClassName2,
					},
					corev1.DNSDefault,
					&corev1.PodDNSConfig{
						Nameservers: []string{dnsNameserver},
					},
					true,
					&kubevirtv1.Memory{
						Hugepages: &kubevirtv1.Hugepages{
							PageSize: "2Mi",
						},
					},
					&kubevirtv1.CPU{
						Cores:                 uint32(1),
						Sockets:               uint32(2),
						Threads:               uint32(1),
						DedicatedCPUPlacement: true,
					},
				)

				machineClass3 := generateMachineClass(
					machineClassTemplate,
					machineClassName3,
					"8Gi",
					"local-3",
					&kubevirtv1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("8192Mi"),
						},
					},
					map[string]string{
						"mcm.gardener.cloud/cluster":      namespace,
						"mcm.gardener.cloud/role":         "node",
						"mcm.gardener.cloud/machineclass": machineClassName3,
					},
					"",
					nil,
					false,
					nil,
					nil,
				)

				chartApplier.
					EXPECT().
					Apply(
						context.TODO(),
						filepath.Join(kubevirt.InternalChartsPath, "machine-class"),
						namespace,
						"machine-class",
						kubernetes.Values(map[string]interface{}{"machineClasses": []map[string]interface{}{
							machineClass1,
							machineClass2,
							machineClass3,
						}}),
					).
					Return(nil)

				generateKubeVirtDataVolumes(providerClient)

				By("comparing machine classes")
				err := workerDelegate.DeployMachineClasses(context.TODO())
				Expect(err).NotTo(HaveOccurred())

				By("comparing machine images")
				machineImages, err := workerDelegate.GetMachineImages(context.TODO())
				Expect(machineImages).To(Equal(&kubevirtv1alpha1.WorkerStatus{
					TypeMeta: metav1.TypeMeta{
						APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
						Kind:       "WorkerStatus",
					},
					MachineImages: []kubevirtv1alpha1.MachineImage{
						{
							Name:      machineImageName,
							Version:   machineImageVersion,
							SourceURL: ubuntuSourceURL,
						},
					},
				}))
				Expect(err).NotTo(HaveOccurred())

				By("comparing machine deployments")

				machineDeployments := worker.MachineDeployments{
					{
						Name:                 machineDeploymentName1,
						ClassName:            machineClassName1,
						SecretName:           machineClassName1,
						Minimum:              worker.DistributeOverZones(0, minPool1, 2),
						Maximum:              worker.DistributeOverZones(0, maxPool1, 2),
						MaxSurge:             worker.DistributePositiveIntOrPercent(0, maxSurgePool1, 2, maxPool1),
						MaxUnavailable:       worker.DistributePositiveIntOrPercent(0, maxUnavailablePool1, 2, minPool1),
						MachineConfiguration: machineConfiguration,
					},
					{
						Name:                 machineDeploymentName2,
						ClassName:            machineClassName2,
						SecretName:           machineClassName2,
						Minimum:              worker.DistributeOverZones(1, minPool1, 2),
						Maximum:              worker.DistributeOverZones(1, maxPool1, 2),
						MaxSurge:             worker.DistributePositiveIntOrPercent(1, maxSurgePool1, 2, maxPool2),
						MaxUnavailable:       worker.DistributePositiveIntOrPercent(1, maxUnavailablePool1, 2, minPool2),
						MachineConfiguration: machineConfiguration,
					},
					{
						Name:                 machineDeploymentName3,
						ClassName:            machineClassName3,
						SecretName:           machineClassName3,
						Minimum:              worker.DistributeOverZones(0, minPool2, 1),
						Maximum:              worker.DistributeOverZones(0, maxPool2, 1),
						MaxSurge:             worker.DistributePositiveIntOrPercent(0, maxSurgePool2, 1, maxPool2),
						MaxUnavailable:       worker.DistributePositiveIntOrPercent(0, maxUnavailablePool2, 1, minPool2),
						MachineConfiguration: machineConfiguration,
					},
				}
				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(machineDeployments))
			})

			It("should fail when the kubevirt secret cannot be read", func() {
				c.EXPECT().
					Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
					Return(fmt.Errorf("error"))

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should fail when the machine image cannot be found", func() {
				cloudProfileName := "test-profile"
				shootVersion := "1.2.3"

				generateKubeVirtSecret(c)

				imagesOutOfConfig := []kubevirtv1alpha1.MachineImages{
					{
						Name: "ubuntu",
						Versions: []kubevirtv1alpha1.MachineImageVersion{
							{
								Version:   "18.04",
								SourceURL: "https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img",
							},
						},
					},
				}

				By("creating a cluster without images")
				cluster := createCluster(cloudProfileName, shootVersion, imagesOutOfConfig)

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster, dataVolumeManager)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should set expected machineControllerManager settings on machine deployment", func() {
				generateKubeVirtSecret(c)

				testDrainTimeout := metav1.Duration{Duration: 10 * time.Minute}
				testHealthTimeout := metav1.Duration{Duration: 20 * time.Minute}
				testCreationTimeout := metav1.Duration{Duration: 30 * time.Minute}
				testMaxEvictRetries := int32(30)
				testNodeConditions := []string{"ReadonlyFilesystem", "KernelDeadlock", "DiskPressure"}
				w.Spec.Pools[0].MachineControllerManagerSettings = &gardencorev1beta1.MachineControllerManagerSettings{
					MachineDrainTimeout:    &testDrainTimeout,
					MachineCreationTimeout: &testCreationTimeout,
					MachineHealthTimeout:   &testHealthTimeout,
					MaxEvictRetries:        &testMaxEvictRetries,
					NodeConditions:         testNodeConditions,
				}

				workerDelegate, _ = NewWorkerDelegate(common.NewClientContext(c, scheme, decoder), chartApplier, "", w, cluster, dataVolumeManager)

				result, err := workerDelegate.GenerateMachineDeployments(context.TODO())
				resultSettings := result[0].MachineConfiguration
				resultNodeConditions := strings.Join(testNodeConditions, ",")
				Expect(err).NotTo(HaveOccurred())
				Expect(resultSettings.MachineDrainTimeout).To(Equal(&testDrainTimeout))
				Expect(resultSettings.MachineCreationTimeout).To(Equal(&testCreationTimeout))
				Expect(resultSettings.MachineHealthTimeout).To(Equal(&testHealthTimeout))
				Expect(resultSettings.MaxEvictRetries).To(Equal(&testMaxEvictRetries))
				Expect(resultSettings.NodeConditions).To(Equal(&resultNodeConditions))
			})
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

func generateKubeVirtSecret(c *mockclient.MockClient) {
	c.EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&corev1.Secret{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, secret *corev1.Secret) error {
			secret.Data = map[string][]byte{
				kubevirt.KubeconfigSecretKey: []byte(kubeconfig),
			}
			return nil
		}).AnyTimes()
}

func generateKubeVirtDataVolumes(providerClient *mockclient.MockClient) {
	providerClient.
		EXPECT().
		Get(context.TODO(), gomock.Any(), gomock.AssignableToTypeOf(&cdicorev1alpha1.DataVolume{})).
		DoAndReturn(func(_ context.Context, _ client.ObjectKey, dataVolume *cdicorev1alpha1.DataVolume) error {
			dataVolume.Spec = cdicorev1alpha1.DataVolumeSpec{
				PVC: &corev1.PersistentVolumeClaimSpec{
					StorageClassName: pointer.StringPtr("standard"),
					AccessModes: []corev1.PersistentVolumeAccessMode{
						"ReadWriteOnce",
					},
				},
				Source: cdicorev1alpha1.DataVolumeSource{
					HTTP: &cdicorev1alpha1.DataVolumeSourceHTTP{
						URL: "https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img",
					},
				},
			}
			return nil
		}).AnyTimes()

	providerClient.
		EXPECT().
		Update(context.TODO(), gomock.Any(), gomock.Any()).
		AnyTimes()
}

func generateMachineClass(classTemplate map[string]interface{}, name, pvcSize, zone string, resources *kubevirtv1.ResourceRequirements,
	tags map[string]string, dnsPolicy corev1.DNSPolicy, dnsConfig *corev1.PodDNSConfig, disablePreAllocatedDataVolumes bool,
	memory *kubevirtv1.Memory, cpu *kubevirtv1.CPU) map[string]interface{} {
	out := make(map[string]interface{})

	for k, v := range classTemplate {
		out[k] = v
	}

	out["name"] = name
	out["pvcSize"] = resource.MustParse(pvcSize)
	out["zone"] = zone
	out["resources"] = resources
	out["tags"] = tags
	out["dnsPolicy"] = dnsPolicy
	out["dnsConfig"] = dnsConfig
	out["disablePreAllocatedDataVolumes"] = disablePreAllocatedDataVolumes
	out["memory"] = memory
	out["cpu"] = cpu

	return out
}

func createCluster(cloudProfileName, shootVersion string, images []kubevirtv1alpha1.MachineImages) *extensionscontroller.Cluster {
	cloudProfileConfig := &kubevirtv1alpha1.CloudProfileConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
			Kind:       "CloudProfileConfig",
		},
		MachineImages: images,
		MachineTypes: []kubevirtv1alpha1.MachineType{
			{
				Name: "local-1",
				Limits: &kubevirtv1alpha1.ResourcesLimits{
					CPU:    resource.MustParse("500m"),
					Memory: resource.MustParse("8Gi"),
				},
			},
		},
	}
	cloudProfileConfigJSON, _ := json.Marshal(cloudProfileConfig)

	cluster := &extensionscontroller.Cluster{
		CloudProfile: &gardencorev1beta1.CloudProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: cloudProfileName,
			},
			Spec: gardencorev1beta1.CloudProfileSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: cloudProfileConfigJSON,
				},
				Regions: []gardencorev1beta1.Region{
					{
						Name: "local",
						Zones: []gardencorev1beta1.AvailabilityZone{
							{
								Name: "local-1",
							},
							{
								Name: "local-2",
							},
						},
					},
				},
				MachineTypes: []gardencorev1beta1.MachineType{
					{
						Name:   "local-1",
						Memory: resource.MustParse("4096Mi"),
						CPU:    resource.MustParse("2"),
						Storage: &gardencorev1beta1.MachineTypeStorage{
							Class:       "standard",
							StorageSize: resource.MustParse("8Gi"),
							Type:        "DataVolume",
						},
					},
					{
						Name:   "local-2",
						Memory: resource.MustParse("8192Mi"),
						CPU:    resource.MustParse("300m"),
						Storage: &gardencorev1beta1.MachineTypeStorage{
							Class:       "standard",
							StorageSize: resource.MustParse("8Gi"),
							Type:        "DataVolume",
						},
					},
				},
			},
		},
		Shoot: &gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				Region: "",
				Kubernetes: gardencorev1beta1.Kubernetes{
					Version: shootVersion,
				},
			},
		},
	}

	specImages := []gardencorev1beta1.MachineImage{}
	for _, image := range images {
		specImages = append(specImages, gardencorev1beta1.MachineImage{
			Name: image.Name,
			Versions: []gardencorev1beta1.ExpirableVersion{
				{
					Version: image.Versions[0].Version,
				},
			},
		})
	}
	cluster.CloudProfile.Spec.MachineImages = specImages

	return cluster
}

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
