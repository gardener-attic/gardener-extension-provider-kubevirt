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
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardener "github.com/gardener/gardener/pkg/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// actuator is an Actuator that acts upon and updates the status of worker resources.
type actuator struct {
	worker.Actuator

	client            client.Client
	dataVolumeManager kubevirt.DataVolumeManager
	logger            logr.Logger
}

// NewActuator creates a new Actuator.
func NewActuator(workerActuator worker.Actuator, dataVolumeManager kubevirt.DataVolumeManager, logger logr.Logger) worker.Actuator {
	return &actuator{
		Actuator:          workerActuator,
		dataVolumeManager: dataVolumeManager,
		logger:            logger.WithName("kubevirt-worker-actuator"),
	}
}

func (a *actuator) InjectFunc(f inject.Func) error {
	return f(a.Actuator)
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

func (a *actuator) Reconcile(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) error {
	if err := a.Actuator.Reconcile(ctx, worker, cluster); err != nil {
		return errors.Wrap(err, "could not reconcile worker")
	}

	return a.deleteOrphanedDataVolumes(ctx, worker)
}

func (a *actuator) Delete(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) error {
	if err := a.Actuator.Delete(ctx, worker, cluster); err != nil {
		return errors.Wrap(err, "could not reconcile worker deletion")
	}

	return a.deleteOrphanedDataVolumes(ctx, worker)
}

func (a *actuator) deleteOrphanedDataVolumes(ctx context.Context, worker *extensionsv1alpha1.Worker) error {
	// Get the kubeconfig of the provider cluster
	kubeconfig, err := kubevirt.GetKubeConfig(ctx, a.client, worker.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get kubeconfig from worker secret reference")
	}

	// List all machine classes in the shoot namespace
	machineClasses := &machinev1alpha1.MachineClassList{}
	if err := a.client.List(ctx, machineClasses, client.InNamespace(worker.Namespace)); err != nil {
		return errors.Wrapf(err, "could not list machine classes in namespace %q", worker.Namespace)
	}
	machineClassNames := names(machineClasses)

	// Initialize labels
	labels := map[string]string{
		kubevirt.ClusterLabel: worker.Namespace,
	}

	// List all data volumes in the provider cluster
	dataVolumes, err := a.dataVolumeManager.ListDataVolumes(ctx, kubeconfig, labels)
	if err != nil {
		return errors.Wrap(err, "could not list data volumes")
	}

	// Delete data volumes that don't have a matching machine class
	for _, dataVolume := range dataVolumes.Items {
		if !machineClassNames.Has(dataVolume.Name) {
			if err := a.dataVolumeManager.DeleteDataVolume(ctx, kubeconfig, dataVolume.Name); err != nil {
				return errors.Wrapf(err, "could not delete orphaned data volume %q", dataVolume.Name)
			}
		}
	}

	return nil
}

func names(machineClasses *machinev1alpha1.MachineClassList) sets.String {
	set := sets.NewString()
	for _, machineClass := range machineClasses.Items {
		set.Insert(machineClass.Name)
	}
	return set
}

type delegateFactory struct {
	common.RESTConfigContext

	clientFactory     kubevirt.ClientFactory
	dataVolumeManager kubevirt.DataVolumeManager
	logger            logr.Logger
}

func (d *delegateFactory) WorkerDelegate(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) (genericactuator.WorkerDelegate, error) {
	clientset, err := kubernetes.NewForConfig(d.RESTConfig())
	if err != nil {
		return nil, errors.Wrap(err, "could not create clientset from REST config")
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, errors.Wrap(err, "could not get server version")
	}

	seedChartApplier, err := gardener.NewChartApplierForConfig(d.RESTConfig())
	if err != nil {
		return nil, errors.Wrap(err, "could not create chart applier from REST config")
	}

	return NewWorkerDelegate(
		d.ClientContext,
		seedChartApplier,
		serverVersion.GitVersion,
		worker,
		cluster,
		d.clientFactory,
		d.dataVolumeManager,
	)
}

type workerDelegate struct {
	common.ClientContext

	seedChartApplier gardener.ChartApplier
	serverVersion    string

	cloudProfileConfig *apiskubevirt.CloudProfileConfig
	cluster            *extensionscontroller.Cluster
	worker             *extensionsv1alpha1.Worker

	machineClasses      []map[string]interface{}
	machineDeployments  worker.MachineDeployments
	machineImages       []apiskubevirt.MachineImage
	machineClassVolumes map[string]*cdicorev1alpha1.DataVolumeSpec

	clientFactory     kubevirt.ClientFactory
	dataVolumeManager kubevirt.DataVolumeManager
}

// NewWorkerDelegate creates a new context for a worker reconciliation.
func NewWorkerDelegate(
	clientContext common.ClientContext,
	seedChartApplier gardener.ChartApplier,
	serverVersion string,
	worker *extensionsv1alpha1.Worker,
	cluster *extensionscontroller.Cluster,
	clientFactory kubevirt.ClientFactory,
	dataVolumeManager kubevirt.DataVolumeManager,
) (genericactuator.WorkerDelegate, error) {
	var cloudProfileConfig *apiskubevirt.CloudProfileConfig
	var err error
	if cluster != nil {
		cloudProfileConfig, err = helper.GetCloudProfileConfig(cluster.CloudProfile)
		if err != nil {
			return nil, errors.Wrap(err, "could not get CloudProfileConfig from cloud profile")
		}
	}

	return &workerDelegate{
		ClientContext:      clientContext,
		seedChartApplier:   seedChartApplier,
		serverVersion:      serverVersion,
		cloudProfileConfig: cloudProfileConfig,
		cluster:            cluster,
		worker:             worker,
		clientFactory:      clientFactory,
		dataVolumeManager:  dataVolumeManager,
	}, nil
}
