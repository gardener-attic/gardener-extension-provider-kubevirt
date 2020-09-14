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

	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Delete deletes the Worker related resources.
func (a *actuator) Delete(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) error {
	if err := a.Actuator.Delete(ctx, worker, cluster); err != nil {
		return errors.Wrap(err, "could not delete worker resouces")
	}

	kubeconfig, err := kubevirt.GetKubeConfig(ctx, a.client, worker.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get kubevirt kubeconfig from worker secret ref")
	}

	machineClasses := &machinev1alpha1.MachineClassList{}
	if err := a.client.List(ctx, machineClasses, client.InNamespace(worker.Namespace)); err != nil {
		return errors.Wrapf(err, "could not list machine classes in namespace %s", worker.Namespace)
	}

	return a.deleteDataVolumes(ctx, kubeconfig, worker, machineClasses)
}

func (a *actuator) deleteDataVolumes(ctx context.Context, kubeconfig []byte, worker *extensionsv1alpha1.Worker,
	machineClasses *machinev1alpha1.MachineClassList) error {
	dataVolumeLabels := map[string]string{
		clusterLabel: worker.Namespace,
	}

	dataVolumes, err := a.dataVolumeManager.ListDataVolumes(ctx, kubeconfig, client.MatchingLabels(dataVolumeLabels))
	if err != nil {
		return errors.Wrapf(err, "could not list data volumes for namespace %s", worker.Namespace)
	}

	for _, dataVolume := range dataVolumes.Items {
		if !machineClassSet(machineClasses).Has(dataVolume.Name) {
			if err := a.dataVolumeManager.DeleteDataVolume(ctx, kubeconfig, dataVolume.Name); err != nil {
				return errors.Wrapf(err, "could not delete data volume %s", dataVolume.Name)
			}
		}
	}

	return nil
}

func machineClassSet(machineClasses *machinev1alpha1.MachineClassList) sets.String {
	machineClassSet := sets.NewString()
	for _, machineClass := range machineClasses.Items {
		machineClassSet.Insert(machineClass.Name)
	}

	return machineClassSet
}
