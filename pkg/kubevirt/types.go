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

package kubevirt

import (
	"path/filepath"
)

const (
	// Type is the type of resources managed by the KubeVirt controllers.
	Type = "kubevirt"

	// Name is the name of the KubeVirt provider controller.
	Name = "provider-kubevirt"

	// ClusterLabel is the label to put on provider cluster objects that identifies the shoot cluster.
	ClusterLabel = "kubevirt.provider.extensions.gardener.cloud/cluster"

	// CloudControllerManagerImageName is the name of the cloud-controller-manager image.
	CloudControllerManagerImageName = "cloud-controller-manager"

	// Kubeconfig is the field in a secret where the kubevirt kubeconfig is stored in.
	Kubeconfig = "kubeconfig"

	// CloudProviderConfigName is the name of the secret containing the cloud provider config.
	CloudProviderConfigName = "cloud-provider-config"
	// CloudControllerManagerName is a constant for the name of the cloud-controller-manager.
	CloudControllerManagerName = "cloud-controller-manager"
	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// MCMProviderKubeVirtImageName is the name of the KubeVirt provider plugin image.
	MCMProviderKubeVirtImageName = "machine-controller-manager-provider-kubevirt"
	// MachineControllerManagerMonitoringConfigName is the name of the ConfigMap containing monitoring stack configurations for machine-controller-manager.
	MachineControllerManagerMonitoringConfigName = "machine-controller-manager-monitoring-config"
	// MachineControllerManagerVpaName is the name of the VerticalPodAutoscaler of the machine-controller-manager deployment.
	MachineControllerManagerVpaName = "machine-controller-manager-vpa"

	// Root disk name
	RootDiskName = "root-disk"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
