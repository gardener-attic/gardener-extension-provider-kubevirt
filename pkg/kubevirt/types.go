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
	// Name is the name of the KubeVirt provider controller.
	Name = "provider-kubevirt"

	// CloudControllerManagerImageName is the name of the cloud-controller-manager image.
	CloudControllerManagerImageName = "cloud-controller-manager"

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

	// CSIAttacherImageName is the name of the CSI attacher image.
	CSIAttacherImageName = "csi-attacher"
	// CSINodeDriverRegistrarImageName is the name of the CSI driver registrar image.
	CSINodeDriverRegistrarImageName = "csi-node-driver-registrar"
	// CSIProvisionerImageName is the name of the CSI provisioner image.
	CSIProvisionerImageName = "csi-provisioner"
	// CSIResizerImageName is the name of the CSI resizer image.
	CSIResizerImageName = "csi-resizer"
	// CSILivenessProbeImageName is the name of the liveness-probe image.
	CSILivenessProbeImageName = "csi-liverness-probe"
	// CSIHostPathDriverImageName is the name of the hostpath image.
	CSIHostPathDriverImageName = "csi-hostpath-driver"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
)
