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

package app

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/helper"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/install"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/cmd"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/controlplane"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/healthcheck"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/infrastructure"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/controller/worker"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/webhook/controlplaneexposure"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	controlerworker "github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/util"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/spf13/cobra"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	cdicorev1alpha1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &controllercmd.RESTOptions{}
		mgrOpts  = &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(kubevirt.Name),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			WebhookServerPort:       443,
		}
		configFileOpts = &cmd.ConfigOptions{}

		// options for the infrastructure controller
		infraCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		reconcileOpts = &controllercmd.ReconcilerOptions{}

		// options for the health care controller
		healthCheckCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the control plane controller
		controlPlaneCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the worker controller
		workerCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		workerReconcileOpts = &controlerworker.Options{
			DeployCRDs: true,
		}
		workerCtrlOptsUnprefixed = controllercmd.NewOptionAggregator(workerCtrlOpts, workerReconcileOpts)

		// options for the webhook server
		webhookServerOptions = &webhookcmd.ServerOptions{
			Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
		}

		controllerSwitches = cmd.ControllerSwitchOptions()
		webhookSwitches    = cmd.WebhookSwitchOptions()
		webhookOptions     = webhookcmd.NewAddToManagerOptions(kubevirt.Name, webhookServerOptions, webhookSwitches)

		aggOption = controllercmd.NewOptionAggregator(
			restOpts,
			mgrOpts,
			controllercmd.PrefixOption("controlplane-", controlPlaneCtrlOpts),
			controllercmd.PrefixOption("infrastructure-", infraCtrlOpts),
			controllercmd.PrefixOption("worker-", &workerCtrlOptsUnprefixed),
			controllercmd.PrefixOption("healthcheck-", healthCheckCtrlOpts),
			controllerSwitches,
			configFileOpts,
			reconcileOpts,
			webhookOptions,
		)
	)

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", kubevirt.Name),

		Run: func(cmd *cobra.Command, args []string) {
			if err := aggOption.Complete(); err != nil {
				controllercmd.LogErrAndExit(err, "Error completing options")
			}

			util.ApplyClientConnectionConfigurationToRESTConfig(configFileOpts.Completed().Config.ClientConnection, restOpts.Completed().Config)
			if workerReconcileOpts.Completed().DeployCRDs {
				if err := controlerworker.ApplyMachineResourcesForConfig(ctx, restOpts.Completed().Config); err != nil {
					controllercmd.LogErrAndExit(err, "Error ensuring the machine CRDs")
				}
			}

			// TODO: remove the call and function once the MachineClass CRD deployment is fixed in Gardener
			if err := helper.ApplyMachineClassCRDs(ctx, restOpts.Completed().Config); err != nil {
				controllercmd.LogErrAndExit(err, "Error ensuring the machine class CRD")
			}

			mgr, err := manager.New(restOpts.Completed().Config, mgrOpts.Completed().Options())
			if err != nil {
				controllercmd.LogErrAndExit(err, "Could not instantiate manager")
			}

			scheme := mgr.GetScheme()
			if err := controller.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := install.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := druidv1alpha1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := autoscalingv1beta2.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := machinev1alpha1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := networkv1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := cdicorev1alpha1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := apiextensionsv1beta1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			// add common meta types to schema for controller-runtime to use v1.ListOptions
			metav1.AddToGroupVersion(scheme, machinev1alpha1.SchemeGroupVersion)

			// apply config options
			configFileOpts.Completed().ApplyETCDStorage(&controlplaneexposure.DefaultAddOptions.ETCDStorage)
			configFileOpts.Completed().ApplyHealthCheckConfig(&healthcheck.DefaultAddOptions.HealthCheckConfig)

			// apply controller options
			healthCheckCtrlOpts.Completed().Apply(&healthcheck.DefaultAddOptions.Controller)
			controlPlaneCtrlOpts.Completed().Apply(&controlplane.DefaultAddOptions.Controller)
			infraCtrlOpts.Completed().Apply(&infrastructure.DefaultAddOptions.Controller)
			workerCtrlOpts.Completed().Apply(&worker.DefaultAddOptions.Controller)

			// apply reconciliation options
			reconcileOpts.Completed().Apply(&infrastructure.DefaultAddOptions.IgnoreOperationAnnotation)
			reconcileOpts.Completed().Apply(&controlplane.DefaultAddOptions.IgnoreOperationAnnotation)
			reconcileOpts.Completed().Apply(&worker.DefaultAddOptions.IgnoreOperationAnnotation)

			if _, _, err := webhookOptions.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add webhooks to manager")
			}

			if err := controllerSwitches.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add controllers to manager")
			}

			if err := mgr.Start(ctx); err != nil {
				controllercmd.LogErrAndExit(err, "Error running manager")
			}
		},
	}

	aggOption.AddFlags(cmd.Flags())

	return cmd
}
