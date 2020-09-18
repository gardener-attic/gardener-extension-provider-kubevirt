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
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/pkg/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// clusterLabel is the label to put on a NetworkAttachmentDefinition object that identifies its cluster.
	clusterLabel = "kubevirt.provider.extensions.gardener.cloud/cluster"
	// nadCRDName is the NetworkAttachmentDefinition CRD name
	nadCRDName = "network-attachment-definitions.k8s.cni.cncf.io"
)

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

	// Get a client and a namespace for the provider cluster from the kubeconfig
	providerClient, namespace, err := kubevirt.GetClient(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "could not get client from kubeconfig")
	}

	var networks []kubevirtv1alpha1.NetworkStatus

	// Initialize labels
	nadLabels := map[string]string{
		clusterLabel: infra.Namespace,
	}

	// Create or update tenant networks
	for _, tenantNetwork := range config.Networks.TenantNetworks {
		// Determine NetworkAttachmentDefinition name
		name := fmt.Sprintf("%s-%s", infra.Namespace, tenantNetwork.Name)

		// Create or update the NetworkAttachmentDefinition of the tenant network in the provider cluster
		a.logger.Info("Creating or updating NetworkAttachmentDefinition", "name", name, "namespace", namespace)
		nad := &networkv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    nadLabels,
			},
		}
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, providerClient, nad, func() error {
				nad.Spec.Config = tenantNetwork.Config
				return nil
			})
			return err
		}); err != nil {
			return errors.Wrapf(err, "could not create or update NetworkAttachmentDefinition '%s'", kutil.ObjectName(nad))
		}

		// Add the tenant network to the list of networks
		networks = append(networks, kubevirtv1alpha1.NetworkStatus{
			Name:    kutil.ObjectName(nad),
			Default: tenantNetwork.Default,
		})
	}

	// Delete old tenant networks
	// Check if the NetworkAttachmentDefinition CRD exists
	var crdErr error
	if crdErr = providerClient.Get(ctx, kutil.Key("", nadCRDName), &apiextensionsv1beta1.CustomResourceDefinition{}); crdErr != nil && !apierrors.IsNotFound(crdErr) {
		return errors.Wrapf(err, "could not get CRD '%v'", nadCRDName)
	}
	if crdErr == nil {
		// List all tenant networks in namespace
		nadList := &networkv1.NetworkAttachmentDefinitionList{}
		if err := providerClient.List(ctx, nadList, client.InNamespace(namespace), client.MatchingLabels(nadLabels)); err != nil {
			return errors.Wrap(err, "could not list NetworkAttachmentDefinitions")
		}
		for _, nad := range nadList.Items {
			// If the network is no longer listed in the InfrastructureConfig, delete it
			if !containsNetworkWithName(networks, kutil.ObjectName(&nad)) {
				// Delete the NetworkAttachmentDefinition of the tenant network in the provider cluster
				a.logger.Info("Deleting NetworkAttachmentDefinition", "name", nad.Name, "namespace", nad.Namespace)
				if err := client.IgnoreNotFound(providerClient.Delete(ctx, &nad)); err != nil {
					return errors.Wrapf(err, "could not delete NetworkAttachmentDefinition '%s'", kutil.ObjectName(&nad))
				}
			}
		}
	}

	// Check shared networks
	for _, sharedNetwork := range config.Networks.SharedNetworks {
		// Determine NetworkAttachmentDefinition namespace
		ns := sharedNetwork.Namespace
		if ns == "" {
			ns = namespace
		}

		// Get the NetworkAttachmentDefinition of the shared network from the provider cluster
		a.logger.Info("Getting NetworkAttachmentDefinition", "name", sharedNetwork.Name, "namespace", ns)
		nadKey := kutil.Key(ns, sharedNetwork.Name)
		nad := &networkv1.NetworkAttachmentDefinition{}
		if err := providerClient.Get(ctx, nadKey, nad); err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Wrapf(err, "NetworkAttachmentDefinition '%v' not found", nadKey)
			}
			return errors.Wrapf(err, "could not get NetworkAttachmentDefinition '%v'", nadKey)
		}

		// Add the full NetworkAttachmentDefinition name to the list of networks
		networks = append(networks, kubevirtv1alpha1.NetworkStatus{
			Name: kutil.ObjectName(nad),
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

func containsNetworkWithName(networks []kubevirtv1alpha1.NetworkStatus, name string) bool {
	for _, networkStatus := range networks {
		if networkStatus.Name == name {
			return true
		}
	}
	return false
}
