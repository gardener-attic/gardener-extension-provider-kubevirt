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
	"context"
	"sync"

	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/pkg/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// nadCRDName is the NetworkAttachmentDefinition CRD name
	nadCRDName = "network-attachment-definitions.k8s.cni.cncf.io"
)

// NetworkManager manages Multus NetworkAttachmentDefinitions.
type NetworkManager interface {
	// CreateOrUpdateNetworkAttachmentDefinition creates or updates a NetworkAttachmentDefinition with the given name, labels, and config.
	CreateOrUpdateNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, config string) (*networkv1.NetworkAttachmentDefinition, error)
	// DeleteNetworkAttachmentDefinition deletes the NetworkAttachmentDefinition with the given name.
	DeleteNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name string) error
	// GetNetworkAttachmentDefinition retrieves the NetworkAttachmentDefinition with the given name and namespace.
	GetNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name, namespace string) (*networkv1.NetworkAttachmentDefinition, error)
	// ListNetworkAttachmentDefinitions lists all NetworkAttachmentDefinitions with the given labels.
	ListNetworkAttachmentDefinitions(ctx context.Context, kubeconfig []byte, labels map[string]string) (*networkv1.NetworkAttachmentDefinitionList, error)
}

type networkManager struct {
	clientFactory ClientFactory
	logger        logr.Logger
	clients       map[string]*getClientResult
	mx            sync.RWMutex
}

// NewNetworkManager creates a new NetworkManager with the given client factory and logger.
func NewNetworkManager(clientFactory ClientFactory, logger logr.Logger) NetworkManager {
	return &networkManager{
		clientFactory: clientFactory,
		logger:        logger.WithName("kubevirt-network-manager"),
		clients:       make(map[string]*getClientResult),
	}
}

// CreateOrUpdateNetworkAttachmentDefinition creates or updates a NetworkAttachmentDefinition with the given name, labels, and config.
func (n *networkManager) CreateOrUpdateNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, config string) (*networkv1.NetworkAttachmentDefinition, error) {
	c, namespace, err := n.getClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create or update the NetworkAttachmentDefinition in the provider cluster
	n.logger.Info("Creating or updating NetworkAttachmentDefinition", "name", name, "namespace", namespace)
	nad := &networkv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := controllerutil.CreateOrUpdate(ctx, c, nad, func() error {
			nad.Labels = labels
			nad.Spec.Config = config
			return nil
		})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "could not create or update NetworkAttachmentDefinition %q", kutil.ObjectName(nad))
	}

	return nad, nil
}

// DeleteNetworkAttachmentDefinition deletes the NetworkAttachmentDefinition with the given name.
func (n *networkManager) DeleteNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name string) error {
	c, namespace, err := n.getClient(kubeconfig)
	if err != nil {
		return err
	}

	// Delete the NetworkAttachmentDefinition in the provider cluster
	n.logger.Info("Deleting NetworkAttachmentDefinition", "name", name, "namespace", namespace)
	nad := &networkv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := client.IgnoreNotFound(c.Delete(ctx, nad)); err != nil {
		return errors.Wrapf(err, "could not delete NetworkAttachmentDefinition %q", kutil.ObjectName(nad))
	}

	return nil
}

// GetNetworkAttachmentDefinition retrieves the NetworkAttachmentDefinition with the given name and namespace.
func (n *networkManager) GetNetworkAttachmentDefinition(ctx context.Context, kubeconfig []byte, name, namespace string) (*networkv1.NetworkAttachmentDefinition, error) {
	c, kubeconfigNamespace, err := n.getClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Determine NetworkAttachmentDefinition namespace
	if namespace == "" {
		namespace = kubeconfigNamespace
	}

	// Get the NetworkAttachmentDefinition from the provider cluster
	n.logger.Info("Getting NetworkAttachmentDefinition", "name", name, "namespace", namespace)
	nadKey := kutil.Key(namespace, name)
	nad := &networkv1.NetworkAttachmentDefinition{}
	if err := c.Get(ctx, nadKey, nad); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "NetworkAttachmentDefinition '%v' not found", nadKey)
		}
		return nil, errors.Wrapf(err, "could not get NetworkAttachmentDefinition '%v'", nadKey)
	}

	return nad, nil
}

// ListNetworkAttachmentDefinitions lists all NetworkAttachmentDefinitions with the given labels.
func (n *networkManager) ListNetworkAttachmentDefinitions(ctx context.Context, kubeconfig []byte, labels map[string]string) (*networkv1.NetworkAttachmentDefinitionList, error) {
	c, namespace, err := n.getClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Check if the NetworkAttachmentDefinition CRD exists
	nadList := &networkv1.NetworkAttachmentDefinitionList{}
	if err := c.Get(ctx, kutil.Key("", nadCRDName), &apiextensionsv1beta1.CustomResourceDefinition{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nadList, nil
		}
		return nil, errors.Wrapf(err, "could not get CRD %q", nadCRDName)
	}

	// List all NetworkAttachmentDefinitions in the provider cluster matching the given namespace and labels
	n.logger.Info("Listing NetworkAttachmentDefinitions", "namespace", namespace, "labels", labels)
	if err := c.List(ctx, nadList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return nil, errors.Wrap(err, "could not list NetworkAttachmentDefinitions")
	}

	return nadList, nil
}

func (n *networkManager) getClient(kubeconfig []byte) (client.Client, string, error) {
	return getClient(kubeconfig, n.clientFactory, n.clients, &n.mx)
}
