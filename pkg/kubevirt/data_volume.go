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

	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClientFactory creates a client from the kubeconfig saved in the "kubeconfig" field of the given secret.
type ClientFactory interface {
	// GetClient creates a client from the kubeconfig saved in the "kubeconfig" field of the given secret.
	// It also returns the namespace of the kubeconfig's current context.
	GetClient(kubeconfig []byte) (client.Client, string, error)
}

// ClientFactoryFunc is a function that implements ClientFactory.
type ClientFactoryFunc func(kubeconfig []byte) (client.Client, string, error)

// GetClient creates a client from the kubeconfig saved in the "kubeconfig" field of the given secret.
// It also returns the namespace of the kubeconfig's current context.
func (f ClientFactoryFunc) GetClient(kubeconfig []byte) (client.Client, string, error) {
	return f(kubeconfig)
}

// DataVolumeManager manages Kubevirt DataVolume operations.
type DataVolumeManager interface {
	// CreateOrUpdateDataVolume creates a new kubevirt Data Volume from the data volume specs and in the passed namespace.
	CreateOrUpdateDataVolume(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, dataVolumeSpec cdicorev1alpha1.DataVolumeSpec) error
	// GetDataVolume fetches the specified volume by the passed name and namespace and return DataVolumeNotFoundError error in case of
	// not found object error.
	GetDataVolume(ctx context.Context, kubeconfig []byte, name string) (*cdicorev1alpha1.DataVolume, error)
	// ListDataVolumes lists all the Data Volumes which exists in the passed namespace.
	ListDataVolumes(ctx context.Context, kubeconfig []byte, listOpts ...client.ListOption) (*cdicorev1alpha1.DataVolumeList, error)
	// DeleteDataVolume delete the DataVolume based on the passed name and namespace.
	DeleteDataVolume(ctx context.Context, kubeconfig []byte, name string) error
}

type defaultDataVolumeManager struct {
	client ClientFactory
	logger logr.Logger
}

// NewDefaultDataVolumeManager creates a new default manager with a k8s controller-runtime based on the sent kubecomfig.
func NewDefaultDataVolumeManager(client ClientFactory) (DataVolumeManager, error) {
	return &defaultDataVolumeManager{
		client: client,
		logger: log.Log.WithName("kubevirt-data-volume-manager"),
	}, nil
}

// CreateOrUpdateDataVolume creates a new kubevirt Data Volume from the data volume specs and in the passed namespace.
func (d *defaultDataVolumeManager) CreateOrUpdateDataVolume(ctx context.Context, kubeconfig []byte, name string, labels map[string]string, dataVolumeSpec cdicorev1alpha1.DataVolumeSpec) error {
	c, namespace, err := d.client.GetClient(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "could not create kubevirt client")
	}

	dataVolume := &cdicorev1alpha1.DataVolume{}
	dataVolume.Namespace = namespace
	dataVolume.Name = name

	_, err = controllerutil.CreateOrUpdate(ctx, c, dataVolume, func() error {
		dataVolume.Labels = labels
		dataVolume.Spec = dataVolumeSpec
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "could not create or update DataVolume '%s'", kutil.ObjectName(dataVolume))
	}

	return nil
}

// GetDataVolume fetches the specified volume by the passed name and namespace and return DataVolumeNotFoundError error in case of
// not found object error.
func (d *defaultDataVolumeManager) GetDataVolume(ctx context.Context, kubeconfig []byte, name string) (*cdicorev1alpha1.DataVolume, error) {
	c, namespace, err := d.client.GetClient(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not create kubevirt client")
	}

	dataVolume := &cdicorev1alpha1.DataVolume{}
	if err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, dataVolume); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "could not get DataVolume: %s", name)
	}

	return dataVolume, nil
}

// ListDataVolumes lists all the Data Volumes which exists in the passed namespace.
func (d *defaultDataVolumeManager) ListDataVolumes(ctx context.Context, kubeconfig []byte, listOpts ...client.ListOption) (*cdicorev1alpha1.DataVolumeList, error) {
	c, namespace, err := d.client.GetClient(kubeconfig)

	if err != nil {
		return nil, errors.Wrap(err, "could not create kubevirt client")
	}

	dvList := cdicorev1alpha1.DataVolumeList{}
	if err := c.List(ctx, &dvList, listOpts...); err != nil {
		return nil, errors.Wrapf(err, "could not list DataVolumes in namespace %s", namespace)
	}

	if len(dvList.Items) == 0 {
		d.logger.V(2).Info("namespace %s has no data volumes", namespace)
		return nil, nil
	}

	return &dvList, nil
}

// DeleteDataVolume delete the DataVolume based on the passed name and namespace.
func (d *defaultDataVolumeManager) DeleteDataVolume(ctx context.Context, kubeconfig []byte, name string) error {
	c, namespace, err := d.client.GetClient(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "could not create kubevirt client")
	}

	dv := &cdicorev1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	return client.IgnoreNotFound(c.Delete(ctx, dv))
}
