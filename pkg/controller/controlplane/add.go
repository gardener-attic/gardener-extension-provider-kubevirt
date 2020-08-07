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

package controlplane

import (
	"context"

	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the KubeVirt controlplane controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
	// GardenId is the Gardener garden identity
	GardenId string
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) error {
	return controlplane.Add(mgr, controlplane.AddArgs{
		Actuator:          NewActuator(opts.GardenId),
		ControllerOptions: opts.Controller,
		Predicates:        controlplane.DefaultPredicates(opts.IgnoreOperationAnnotation),
		Type:              kubevirt.Type,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(gardenID string) controlplane.Actuator {
	return &actuator{
		logger:   log.Log.WithName("infrastructure-actuator"),
		gardenID: gardenID,
	}
}

type actuator struct {
	common.ChartRendererContext

	logger   logr.Logger
	gardenID string
}

func (a *actuator) Reconcile(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (bool, error) {
	a.logger.Info("control-plane reconciled")
	// TODO: install kubevirt-cloud-controller-manager here and related components, the genericActuator might be used here
	return false, nil
}

// Delete deletes the ControlPlane.
func (a *actuator) Delete(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) error {
	return nil
}

// Restore restores the ControlPlane.
func (a *actuator) Restore(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) (bool, error) {
	return false, nil
}

// Migrate migrates the ControlPlane.
func (a *actuator) Migrate(context.Context, *extensionsv1alpha1.ControlPlane, *extensionscontroller.Cluster) error {
	return nil
}
