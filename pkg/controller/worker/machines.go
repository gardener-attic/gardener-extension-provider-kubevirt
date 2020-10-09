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

package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	kubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

// MachineClassKind yields the name of the KubeVirt machine class.
func (w *workerDelegate) MachineClassKind() string {
	return "MachineClass"
}

// MachineClassList yields a newly initialized KubeVirtMachineClassList object.
func (w *workerDelegate) MachineClassList() runtime.Object {
	return &machinev1alpha1.MachineClassList{}
}

// DeployMachineClasses generates and creates the KubeVirt specific machine classes.
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	if err := w.seedChartApplier.Apply(
		ctx, filepath.Join(kubevirt.InternalChartsPath, "machine-class"), w.worker.Namespace, "machine-class",
		kubernetes.Values(map[string]interface{}{"machineClasses": w.machineClasses}),
	); err != nil {
		return errors.Wrap(err, "could not apply machine-class chart")
	}

	if err := w.createOrUpdateMachineClassVolumes(ctx); err != nil {
		return err
	}

	return nil
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineDeployments  = worker.MachineDeployments{}
		machineClasses      []map[string]interface{}
		machineImages       []apiskubevirt.MachineImage
		machineClassVolumes = make(map[string]*cdicorev1alpha1.DataVolumeSpec)
	)

	kubeconfig, err := kubevirt.GetKubeConfig(ctx, w.Client(), w.worker.Spec.SecretRef)
	if err != nil {
		return err
	}

	// Get a client and a namespace for the provider cluster from the kubeconfig
	_, namespace, err := w.clientFactory.GetClient(kubeconfig)
	if err != nil {
		return err
	}

	infrastructureStatus, err := helper.GetInfrastructureStatus(w.worker)
	if err != nil {
		return err
	}

	infrastructureStatusV1alpha1 := &kubevirtv1alpha1.InfrastructureStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
			Kind:       "InfrastructureStatus",
		},
	}
	if err := w.Scheme().Convert(infrastructureStatus, infrastructureStatusV1alpha1, nil); err != nil {
		return err
	}

	var networksData []string
	for _, network := range infrastructureStatus.Networks {
		networksData = append(networksData, network.Name, strconv.FormatBool(network.Default), network.SHA)
	}

	if len(w.worker.Spec.SSHPublicKey) == 0 {
		return fmt.Errorf("missing sshPublicKey in worker")
	}

	for _, pool := range w.worker.Spec.Pools {
		zoneLen := int32(len(pool.Zones))

		workerConfig, err := helper.GetWorkerConfig(&pool)
		if err != nil {
			return err
		}

		machineType, err := w.getMachineType(pool.MachineType)
		if err != nil {
			return err
		}

		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster, networksData...)
		if err != nil {
			return err
		}

		imageSourceURL, err := w.getMachineImageURL(pool.MachineImage.Name, pool.MachineImage.Version)
		if err != nil {
			return err
		}
		machineImages = appendMachineImage(machineImages, apiskubevirt.MachineImage{
			Name:      pool.MachineImage.Name,
			Version:   pool.MachineImage.Version,
			SourceURL: imageSourceURL,
		})

		resourceRequirements := &kubevirtv1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    machineType.CPU,
				corev1.ResourceMemory: machineType.Memory,
			},
			OvercommitGuestOverhead: workerConfig.OvercommitGuestOverhead,
		}

		if mt := w.getMachineTypesExtension(machineType.Name); mt != nil {
			if mt.Limits != nil {
				resourceRequirements.Limits = corev1.ResourceList{
					corev1.ResourceCPU:    mt.Limits.CPU,
					corev1.ResourceMemory: mt.Limits.Memory,
				}
			}
		}

		// Get root volume storage class name and size
		var rootVolumeClassName string
		var rootVolumeSize resource.Quantity
		switch {
		case pool.Volume != nil:
			rootVolumeClassName, rootVolumeSize, err = w.getStorageClassNameAndSize(*pool.Volume.Type, pool.Volume.Size)
			if err != nil {
				return err
			}
		case machineType.Storage != nil:
			rootVolumeClassName, rootVolumeSize = machineType.Storage.Class, machineType.Storage.StorageSize
		default:
			return fmt.Errorf("missing root volume parameters")
		}

		// Build additional volumes
		var additionalVolumes []map[string]interface{}
		for _, volume := range pool.DataVolumes {
			className, size, err := w.getStorageClassNameAndSize(*volume.Type, volume.Size)
			if err != nil {
				return err
			}
			additionalVolumes = append(additionalVolumes, map[string]interface{}{
				"dataVolume": buildDataVolumeSpecWithBlankSource(className, size),
			})
		}

		for zoneIndex, zone := range pool.Zones {
			zoneIdx := int32(zoneIndex)

			deploymentName := fmt.Sprintf("%s-%s-z%d", w.worker.Namespace, pool.Name, zoneIndex+1)
			className := fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)

			// Build root volume and machine class volume
			var rootVolume *cdicorev1alpha1.DataVolumeSpec
			if !workerConfig.DisablePreAllocatedDataVolumes {
				rootVolume = buildDataVolumeSpecWithPVCSource(rootVolumeClassName, rootVolumeSize, namespace, className)
				machineClassVolumes[className] = buildDataVolumeSpecWithHTTPSource(rootVolumeClassName, rootVolumeSize, imageSourceURL)
			} else {
				rootVolume = buildDataVolumeSpecWithHTTPSource(rootVolumeClassName, rootVolumeSize, imageSourceURL)
			}

			machineClasses = append(machineClasses, map[string]interface{}{
				"name":              className,
				"resources":         resourceRequirements,
				"rootVolume":        rootVolume,
				"additionalVolumes": additionalVolumes,
				"sshKeys":           []string{string(w.worker.Spec.SSHPublicKey)},
				"networks":          infrastructureStatusV1alpha1.Networks,
				"region":            w.worker.Spec.Region,
				"zone":              zone,
				"tags": map[string]string{
					"mcm.gardener.cloud/cluster":      w.worker.Namespace,
					"mcm.gardener.cloud/role":         "node",
					"mcm.gardener.cloud/machineclass": className,
				},
				"secret": map[string]interface{}{
					"cloudConfig": string(pool.UserData),
					"kubeconfig":  string(kubeconfig),
				},
				"dnsPolicy": workerConfig.DNSPolicy,
				"dnsConfig": workerConfig.DNSConfig,
				"memory":    workerConfig.Memory,
				"cpu":       workerConfig.CPU,
			})

			machineDeployments = append(machineDeployments, worker.MachineDeployment{
				Name:                 deploymentName,
				ClassName:            className,
				SecretName:           className,
				Minimum:              worker.DistributeOverZones(zoneIdx, pool.Minimum, zoneLen),
				Maximum:              worker.DistributeOverZones(zoneIdx, pool.Maximum, zoneLen),
				MaxSurge:             worker.DistributePositiveIntOrPercent(zoneIdx, pool.MaxSurge, zoneLen, pool.Maximum),
				MaxUnavailable:       worker.DistributePositiveIntOrPercent(zoneIdx, pool.MaxUnavailable, zoneLen, pool.Minimum),
				Labels:               pool.Labels,
				Annotations:          pool.Annotations,
				Taints:               pool.Taints,
				MachineConfiguration: genericworkeractuator.ReadMachineConfiguration(pool),
			})
		}
	}

	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses
	w.machineImages = machineImages
	w.machineClassVolumes = machineClassVolumes

	return nil
}

func (w *workerDelegate) getStorageClassNameAndSize(volumeTypeName, volumeSize string) (string, resource.Quantity, error) {
	volumeType, err := w.getVolumeType(volumeTypeName)
	if err != nil {
		return "", resource.Quantity{}, err
	}
	storageSize, err := resource.ParseQuantity(volumeSize)
	if err != nil {
		return "", resource.Quantity{}, err
	}
	return volumeType.Class, storageSize, nil
}

func (w *workerDelegate) getMachineType(name string) (*corev1beta1.MachineType, error) {
	for _, mt := range w.cluster.CloudProfile.Spec.MachineTypes {
		if mt.Name == name {
			return &mt, nil
		}
	}
	return nil, fmt.Errorf("machine type %s not found in cloud profile spec", name)
}

func (w *workerDelegate) getVolumeType(name string) (*corev1beta1.VolumeType, error) {
	for _, vt := range w.cluster.CloudProfile.Spec.VolumeTypes {
		if vt.Name == name {
			return &vt, nil
		}
	}
	return nil, fmt.Errorf("volume type %s not found in cloud profile spec", name)
}

func (w *workerDelegate) createOrUpdateMachineClassVolumes(ctx context.Context) error {
	labels := map[string]string{
		clusterLabel: w.worker.Namespace,
	}

	kubeconfig, err := kubevirt.GetKubeConfig(ctx, w.Client(), w.worker.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get kubeconfig of the kubevirt cluster")
	}

	for name, dataVolumeSpec := range w.machineClassVolumes {
		if err := w.dataVolumeManager.CreateOrUpdateDataVolume(ctx, kubeconfig, name, labels, *dataVolumeSpec); err != nil {
			return errors.Wrapf(err, "could not create data volume for machine class %s", name)
		}
	}

	return nil
}

func buildDataVolumeSpecWithPVCSource(storageClassName string, storageSize resource.Quantity, namespace, name string) *cdicorev1alpha1.DataVolumeSpec {
	return &cdicorev1alpha1.DataVolumeSpec{
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
			StorageClassName: pointer.StringPtr(storageClassName),
		},
		Source: cdicorev1alpha1.DataVolumeSource{
			PVC: &cdicorev1alpha1.DataVolumeSourcePVC{
				Namespace: namespace,
				Name:      name,
			},
		},
	}
}

func buildDataVolumeSpecWithHTTPSource(storageClassName string, storageSize resource.Quantity, url string) *cdicorev1alpha1.DataVolumeSpec {
	return &cdicorev1alpha1.DataVolumeSpec{
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
			StorageClassName: pointer.StringPtr(storageClassName),
		},
		Source: cdicorev1alpha1.DataVolumeSource{
			HTTP: &cdicorev1alpha1.DataVolumeSourceHTTP{
				URL: url,
			},
		},
	}
}

func buildDataVolumeSpecWithBlankSource(storageClassName string, storageSize resource.Quantity) *cdicorev1alpha1.DataVolumeSpec {
	return &cdicorev1alpha1.DataVolumeSpec{
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
			StorageClassName: pointer.StringPtr(storageClassName),
		},
		Source: cdicorev1alpha1.DataVolumeSource{
			Blank: &cdicorev1alpha1.DataVolumeBlankImage{},
		},
	}
}
