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
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

const clusterLabel = "kubevirt.provider.extensions.gardener.cloud/cluster"

var workerLogger = log.Log.WithName("kubevirt-worker-actuator")

// actuator is an Actuator that acts upon and updates the status of worker resources.
type actuator struct {
	worker.Actuator

	dataVolumeManager kubevirt.DataVolumeManager
	client            client.Client
	logger            logr.Logger
}

// NewActuator creates a new Actuator that updates the status of the handled WorkerPoolConfigs.
func NewActuator(workerActuator worker.Actuator, client client.Client, logger logr.Logger,
	dataVolumeManager kubevirt.DataVolumeManager) worker.Actuator {

	return &actuator{
		Actuator:          workerActuator,
		logger:            logger,
		dataVolumeManager: dataVolumeManager,
		client:            client,
	}
}

func (a *actuator) InjectFunc(f inject.Func) error {
	return f(a.Actuator)
}

func (a *actuator) InjectClient(client client.Client) error {
	a.client = client
	return nil
}

type delegateFactory struct {
	logger            logr.Logger
	clientFactory     kubevirt.ClientFactory
	dataVolumeManager kubevirt.DataVolumeManager
	common.RESTConfigContext
}

func (d *delegateFactory) WorkerDelegate(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) (genericactuator.WorkerDelegate, error) {
	clientset, err := kubernetes.NewForConfig(d.RESTConfig())
	if err != nil {
		return nil, err
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	seedChartApplier, err := gardener.NewChartApplierForConfig(d.RESTConfig())
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	return &workerDelegate{
		ClientContext: clientContext,

		seedChartApplier: seedChartApplier,
		serverVersion:    serverVersion,

		cloudProfileConfig: cloudProfileConfig,
		cluster:            cluster,
		worker:             worker,
		clientFactory:      clientFactory,
		dataVolumeManager:  dataVolumeManager,
	}, nil
}
