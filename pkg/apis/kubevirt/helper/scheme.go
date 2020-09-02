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

package helper

import (
	"context"

	api "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/install"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	// Scheme is a scheme with the types relevant for KubeVirt actuators.
	Scheme *runtime.Scheme

	decoder runtime.Decoder
)

func init() {
	Scheme = runtime.NewScheme()
	utilruntime.Must(install.AddToScheme(Scheme))

	// TODO: remove after MachineClass CRD deployment is fixed in gardener
	utilruntime.Must(apiextensionsscheme.AddToScheme(Scheme))

	decoder = serializer.NewCodecFactory(Scheme).UniversalDecoder()
}

// ApplyMachineClassCRDs applies the MachineClass CRD,
// currently, gardener does not apply MachineClass for OOT approach
// this function should be removed once it's fixed in Gardner
func ApplyMachineClassCRDs(ctx context.Context, config *rest.Config) error {
	deletionProtectionLabels := map[string]string{
		common.GardenerDeletionProtected: "true",
	}

	machineClassCRD := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machineclasses.machine.sapcloud.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group: "machine.sapcloud.io",
			Versions: []v1beta1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				},
			},
			Scope: apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:       "MachineClass",
				Plural:     "machineclasses",
				Singular:   "machineclass",
				ShortNames: []string{"cls"},
			},
			Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
				Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
			},
		},
	}

	c, err := client.New(config, client.Options{Scheme: Scheme})
	if err != nil {
		return err
	}

	spec := machineClassCRD.Spec.DeepCopy()
	_, err = controllerutil.CreateOrUpdate(ctx, c, machineClassCRD, func() error {
		machineClassCRD.Labels = utils.MergeStringMaps(machineClassCRD.Labels, deletionProtectionLabels)
		machineClassCRD.Spec = *spec
		return nil
	})

	return err
}

// GetCloudProfileConfig extracts the CloudProfileConfig from the ProviderConfig section of the given CloudProfile.
func GetCloudProfileConfig(cloudProfile *gardencorev1beta1.CloudProfile) (*api.CloudProfileConfig, error) {
	cloudProfileConfig := &api.CloudProfileConfig{}
	if cloudProfile.Spec.ProviderConfig != nil && cloudProfile.Spec.ProviderConfig.Raw != nil {
		if _, _, err := decoder.Decode(cloudProfile.Spec.ProviderConfig.Raw, nil, cloudProfileConfig); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of cloudProfile '%s'", kutil.ObjectName(cloudProfile))
		}
	}
	return cloudProfileConfig, nil
}

// GetInfrastructureConfig extracts the InfrastructureConfig from the ProviderConfig section of the given Infrastructure.
func GetInfrastructureConfig(infra *extensionsv1alpha1.Infrastructure) (*api.InfrastructureConfig, error) {
	config := &api.InfrastructureConfig{}
	if infra.Spec.ProviderConfig != nil && infra.Spec.ProviderConfig.Raw != nil {
		if _, _, err := decoder.Decode(infra.Spec.ProviderConfig.Raw, nil, config); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of infrastructure '%s'", kutil.ObjectName(infra))
		}
	}
	return config, nil
}

// GetControlPlaneConfig extracts the ControlPlaneConfig from the ProviderConfig section of the given ControlPlane.
func GetControlPlaneConfig(cp *extensionsv1alpha1.ControlPlane) (*api.ControlPlaneConfig, error) {
	config := &api.ControlPlaneConfig{}
	if cp.Spec.ProviderConfig != nil && cp.Spec.ProviderConfig.Raw != nil {
		if _, _, err := decoder.Decode(cp.Spec.ProviderConfig.Raw, nil, config); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", kutil.ObjectName(cp))
		}
	}
	return config, nil
}

// GetInfrastructureStatus extracts the InfrastructureStatus from the InfrastructureProviderStatus section of the given Worker.
func GetInfrastructureStatus(w *extensionsv1alpha1.Worker) (*api.InfrastructureStatus, error) {
	status := &api.InfrastructureStatus{}
	if w.Spec.InfrastructureProviderStatus != nil && w.Spec.InfrastructureProviderStatus.Raw != nil {
		if _, _, err := decoder.Decode(w.Spec.InfrastructureProviderStatus.Raw, nil, status); err != nil {
			return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of worker '%s'", kutil.ObjectName(w))
		}
	}
	return status, nil
}
