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

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
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

	dataVolumeLabels := map[string]string{
		clusterLabel: w.worker.Namespace,
	}

	kubeconfig, err := kubevirt.GetKubeConfig(ctx, w.Client(), w.worker.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get the kubeconfig of the kubevirt cluster")
	}

	for _, machineClass := range w.machineClasses {
		var (
			storageClassName = pointer.StringPtr(machineClass["storageClassName"].(string))
			pvcSize          = v1.ResourceList{v1.ResourceStorage: machineClass["pvcSize"].(resource.Quantity)}
			sourceUrl        = machineClass["sourceURL"].(string)
			machineClassName = machineClass["name"].(string)
		)

		dataVolumeSpec := cdicorev1alpha1.DataVolumeSpec{
			PVC: &v1.PersistentVolumeClaimSpec{
				StorageClassName: storageClassName,
				AccessModes: []v1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: v1.ResourceRequirements{
					Requests: pvcSize,
				},
			},
			Source: cdicorev1alpha1.DataVolumeSource{
				HTTP: &cdicorev1alpha1.DataVolumeSourceHTTP{
					URL: sourceUrl,
				},
			},
		}

		if err := w.dataVolumeManager.CreateOrUpdateDataVolume(ctx, kubeconfig, machineClassName, dataVolumeLabels, dataVolumeSpec); err != nil {
			return errors.Wrapf(err, "could not create data volume for machine class %s", machineClassName)
		}
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
		machineDeployments = worker.MachineDeployments{}
		machineClasses     []map[string]interface{}
		machineImages      []apiskubevirt.MachineImage
	)

	kubeconfig, err := kubevirt.GetKubeConfig(ctx, w.Client(), w.worker.Spec.SecretRef)
	if err != nil {
		return err
	}

	infrastructureStatus, err := helper.GetInfrastructureStatus(w.worker)
	if err != nil {
		return err
	}

	if len(w.worker.Spec.SSHPublicKey) == 0 {
		return fmt.Errorf("missing sshPublicKey in worker")
	}

	for _, pool := range w.worker.Spec.Pools {
		// hardcoded for now as we don't support zones yet
		zoneIdx := int32(0)
		zoneLen := int32(1)

		workerConfig, err := helper.GetWorkerConfig(&pool)
		if err != nil {
			return err
		}

		machineType, err := w.getMachineType(pool.MachineType)
		if err != nil {
			return err
		}

		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster)
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

		deploymentName := fmt.Sprintf("%s-%s-z", w.worker.Namespace, pool.Name)
		className := fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)

		var networks []map[string]interface{}
		for _, networkStatus := range infrastructureStatus.Networks {
			networks = append(networks, map[string]interface{}{
				"name":    networkStatus.Name,
				"default": networkStatus.Default,
			})
		}

		machineClasses = append(machineClasses, map[string]interface{}{
			"name":             className,
			"storageClassName": machineType.Storage.Class,
			"pvcSize":          machineType.Storage.StorageSize,
			"sourceURL":        imageSourceURL,
			"cpus":             machineType.CPU,
			"memory":           machineType.Memory,
			"sshKeys":          []string{string(w.worker.Spec.SSHPublicKey)},
			"networks":         networks,
			"region":           w.worker.Spec.Region,
			"zones":            pool.Zones,
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

	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses
	w.machineImages = machineImages

	return nil
}

func (w *workerDelegate) getMachineType(name string) (*corev1beta1.MachineType, error) {
	for _, mt := range w.cluster.CloudProfile.Spec.MachineTypes {
		if mt.Name == name {
			return &mt, nil
		}
	}
	return nil, fmt.Errorf("machine type %s not found in cloud profile spec", name)
}
