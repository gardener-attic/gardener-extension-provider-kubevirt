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

package infrastructure

import (
	"context"
	"fmt"

	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	kubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

type actuator struct {
	common.ChartRendererContext

	networkManager kubevirt.NetworkManager
	logger         logr.Logger
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(networkManager kubevirt.NetworkManager, logger logr.Logger) infrastructure.Actuator {
	return &actuator{
		networkManager: networkManager,
		logger:         logger.WithName("kubevirt-infrastructure-actuator"),
	}
}

func (a *actuator) Reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	// Get InfrastructureConfig from the Infrastructure resource
	config, err := helper.GetInfrastructureConfig(infra)
	if err != nil {
		return errors.Wrap(err, "could not get InfrastructureConfig from infrastructure")
	}

	// Get the kubeconfig of the provider cluster
	kubeconfig, err := kubevirt.GetKubeConfig(ctx, a.Client(), infra.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get kubeconfig from infrastructure secret reference")
	}

	var networks []kubevirtv1alpha1.NetworkStatus

	// Initialize labels
	labels := map[string]string{
		kubevirt.ClusterLabel: infra.Namespace,
	}

	// Create or update tenant networks
	for _, tenantNetwork := range config.Networks.TenantNetworks {
		// Determine NetworkAttachmentDefinition name
		name := fmt.Sprintf("%s-%s", infra.Namespace, tenantNetwork.Name)

		// Create or update the tenant network in the provider cluster
		nad, err := a.networkManager.CreateOrUpdateNetworkAttachmentDefinition(ctx, kubeconfig, name, labels, tenantNetwork.Config)
		if err != nil {
			return errors.Wrapf(err, "could not create or update tenant network %q", name)
		}

		// Add the tenant network to the list of networks
		networks = append(networks, kubevirtv1alpha1.NetworkStatus{
			Name:    kutil.ObjectName(nad),
			Default: tenantNetwork.Default,
			SHA:     utils.ComputeSHA256Hex([]byte(tenantNetwork.Config)),
		})
	}

	// Delete old tenant networks
	// List all tenant networks in the provider cluster
	nadList, err := a.networkManager.ListNetworkAttachmentDefinitions(ctx, kubeconfig, labels)
	if err != nil {
		return errors.Wrap(err, "could not list tenant networks")
	}
	for _, nad := range nadList.Items {
		// If the tenant network is no longer listed in the InfrastructureConfig, delete it
		if !containsNetworkWithName(networks, kutil.ObjectName(&nad)) {
			// Delete the tenant network in the provider cluster
			if err := a.networkManager.DeleteNetworkAttachmentDefinition(ctx, kubeconfig, nad.Name); err != nil {
				return errors.Wrapf(err, "could not delete tenant network %q", nad.Name)
			}
		}
	}

	// Check shared networks
	for _, sharedNetwork := range config.Networks.SharedNetworks {
		// Get the the shared network from the provider cluster
		nad, err := a.networkManager.GetNetworkAttachmentDefinition(ctx, kubeconfig, sharedNetwork.Name, sharedNetwork.Namespace)
		if err != nil {
			return errors.Wrapf(err, "could not get shared network '%s/%s'", sharedNetwork.Namespace, sharedNetwork.Name)
		}

		// Add the full shared network name to the list of networks
		networks = append(networks, kubevirtv1alpha1.NetworkStatus{
			Name: kutil.ObjectName(nad),
			SHA:  utils.ComputeSHA256Hex([]byte(nad.Spec.Config)),
		})
	}

	// Update infrastructure status
	a.logger.Info("Updating infrastructure status")
	status := &kubevirtv1alpha1.InfrastructureStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirtv1alpha1.SchemeGroupVersion.String(),
			Kind:       "InfrastructureStatus",
		},
		Networks: networks,
	}
	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client(), infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
		return nil
	})
}

func (a *actuator) Delete(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	// Get InfrastructureConfig from the Infrastructure resource
	config, err := helper.GetInfrastructureConfig(infra)
	if err != nil {
		return errors.Wrap(err, "could not get InfrastructureConfig from infrastructure")
	}

	// Get the kubeconfig of the provider cluster
	kubeconfig, err := kubevirt.GetKubeConfig(ctx, a.Client(), infra.Spec.SecretRef)
	if err != nil {
		return errors.Wrap(err, "could not get kubeconfig from infrastructure secret reference")
	}

	// Delete tenant networks
	for _, tenantNetwork := range config.Networks.TenantNetworks {
		// Determine NetworkAttachmentDefinition name
		name := fmt.Sprintf("%s-%s", infra.Namespace, tenantNetwork.Name)

		// Delete the tenant network in the provider cluster
		if err := a.networkManager.DeleteNetworkAttachmentDefinition(ctx, kubeconfig, name); err != nil {
			return errors.Wrapf(err, "could not delete tenant network %q", name)
		}
	}

	return nil
}

func (a *actuator) Restore(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	return nil
}

func (a *actuator) Migrate(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	return nil
}

func containsNetworkWithName(networks []kubevirtv1alpha1.NetworkStatus, name string) bool {
	for _, networkStatus := range networks {
		if networkStatus.Name == name {
			return true
		}
	}
	return false
}
