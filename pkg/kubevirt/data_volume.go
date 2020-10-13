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
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DataVolumeManager manages KubeVirt DataVolumes.
type DataVolumeManager interface {
	// CreateOrUpdateDataVolume creates or updates a DataVolume with the given name, labels, and spec.
	CreateOrUpdateDataVolume(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, spec cdicorev1alpha1.DataVolumeSpec) (*cdicorev1alpha1.DataVolume, error)
	// DeleteDataVolume deletes the DataVolume with the given name.
	DeleteDataVolume(ctx context.Context, kubeconfig []byte, name string) error
	// ListDataVolumes lists all DataVolumes matching the given labels.
	ListDataVolumes(ctx context.Context, kubeconfig []byte, labels map[string]string) (*cdicorev1alpha1.DataVolumeList, error)
}

type dataVolumeManager struct {
	clientFactory ClientFactory
	logger        logr.Logger
	clients       map[string]*getClientResult
	mx            sync.RWMutex
}

// NewDataVolumeManager creates a new DataVolumeManager with the given client factory and logger.
func NewDataVolumeManager(clientFactory ClientFactory, logger logr.Logger) DataVolumeManager {
	return &dataVolumeManager{
		clientFactory: clientFactory,
		logger:        logger.WithName("kubevirt-datavolume-manager"),
		clients:       make(map[string]*getClientResult),
	}
}

// CreateOrUpdateDataVolume creates or updates a DataVolume with the given name, labels, and spec.
func (d *dataVolumeManager) CreateOrUpdateDataVolume(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, spec cdicorev1alpha1.DataVolumeSpec) (*cdicorev1alpha1.DataVolume, error) {
	c, namespace, err := d.getClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create or update the DataVolume in the provider cluster
	d.logger.Info("Creating or updating DataVolume", "name", name, "namespace", namespace)
	dv := &cdicorev1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err = controllerutil.CreateOrUpdate(ctx, c, dv, func() error {
			dv.Labels = labels
			dv.Spec = spec
			return nil
		})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "could not create or update DataVolume %q", kutil.ObjectName(dv))
	}

	return dv, nil
}

// DeleteDataVolume deletes the DataVolume with the given name.
func (d *dataVolumeManager) DeleteDataVolume(ctx context.Context, kubeconfig []byte, name string) error {
	c, namespace, err := d.getClient(kubeconfig)
	if err != nil {
		return err
	}

	// Delete the DataVolume in the provider cluster
	d.logger.Info("Deleting DataVolume", "name", name, "namespace", namespace)
	dv := &cdicorev1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := client.IgnoreNotFound(c.Delete(ctx, dv)); err != nil {
		return errors.Wrapf(err, "could not delete DataVolume %q", kutil.ObjectName(dv))
	}

	return nil
}

// ListDataVolumes lists all DataVolumes matching the given labels.
func (d *dataVolumeManager) ListDataVolumes(ctx context.Context, kubeconfig []byte, labels map[string]string) (*cdicorev1alpha1.DataVolumeList, error) {
	c, namespace, err := d.getClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// List all DataVolumes in the provider cluster matching the given namespace and labels
	d.logger.Info("Listing DataVolumes", "namespace", namespace, "labels", labels)
	dvList := cdicorev1alpha1.DataVolumeList{}
	if err := c.List(ctx, &dvList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return nil, errors.Wrapf(err, "could not list DataVolumes in namespace %q", namespace)
	}

	return &dvList, nil
}

func (d *dataVolumeManager) getClient(kubeconfig []byte) (client.Client, string, error) {
	return getClient(kubeconfig, d.clientFactory, d.clients, &d.mx)
}
