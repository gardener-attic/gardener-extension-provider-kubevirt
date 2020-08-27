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
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

	// Get a client and a namespace for the provider cluster from the kubeconfig
	providerClient, namespace, err := kubevirt.GetClient(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "could not get client from kubeconfig")
	}

	// Delete tenant networks
	for _, tenantNetwork := range config.Networks.TenantNetworks {
		// Determine NetworkAttachmentDefinition name
		name := fmt.Sprintf("%s-%s", infra.Namespace, tenantNetwork.Name)

		// Delete the NetworkAttachmentDefinition of the tenant network in the provider cluster
		a.logger.Info("Deleting NetworkAttachmentDefinition", "name", name, "namespace", namespace)
		nad := &networkv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		if err := client.IgnoreNotFound(providerClient.Delete(ctx, nad)); err != nil {
			return errors.Wrapf(err, "could not delete NetworkAttachmentDefinition '%s'", kutil.ObjectName(nad))
		}
	}

	return nil
}
