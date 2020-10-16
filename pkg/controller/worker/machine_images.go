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

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	kubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

// UpdateMachineImagesStatus updates the machine image status
// with the used machine images for the `Worker` resource.
func (w *workerDelegate) UpdateMachineImagesStatus(ctx context.Context) error {
	if w.machineImages == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	// Decode the current worker provider status.
	workerStatus, err := helper.GetWorkerStatus(w.worker)
	if err != nil {
		return errors.Wrap(err, "could not get WorkerStatus from worker")
	}
	workerStatus.MachineImages = w.machineImages

	var workerStatusV1alpha1 = &kubevirtv1alpha1.WorkerStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
			Kind:       "WorkerStatus",
		},
	}

	if err := w.Scheme().Convert(workerStatus, workerStatusV1alpha1, nil); err != nil {
		return errors.Wrap(err, "could not convert WorkerStatus to v1alpha1")
	}

	return controller.TryUpdateStatus(ctx, retry.DefaultBackoff, w.Client(), w.worker, func() error {
		w.worker.Status.ProviderStatus = &runtime.RawExtension{Object: workerStatusV1alpha1}
		return nil
	})
}

func (w *workerDelegate) getMachineImageURL(name, version string) (string, error) {
	if sourceURL, err := helper.FindImageFromCloudProfile(w.cloudProfileConfig, name, version); err == nil {
		return sourceURL, nil
	}

	// Try to look up machine image in worker provider status as it was not found in the cloud profile.
	workerStatus, err := helper.GetWorkerStatus(w.worker)
	if err != nil {
		return "", errors.Wrap(err, "could not get WorkerStatus from worker")
	}
	if machineImage, err := helper.FindMachineImage(workerStatus.MachineImages, name, version); err == nil {
		return machineImage.SourceURL, nil
	}

	return "", worker.ErrorMachineImageNotFound(name, version)
}

func appendMachineImage(machineImages []apiskubevirt.MachineImage, machineImage apiskubevirt.MachineImage) []apiskubevirt.MachineImage {
	if _, err := helper.FindMachineImage(machineImages, machineImage.Name, machineImage.Version); err != nil {
		return append(machineImages, machineImage)
	}
	return machineImages
}
