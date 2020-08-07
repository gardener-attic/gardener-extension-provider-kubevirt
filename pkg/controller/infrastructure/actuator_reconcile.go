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

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
	apiskubevirtv1alpha1 "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt/v1alpha1"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
)

func (a *actuator) Reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	// TODO: here, controller should ensure the network setup
	status := &apiskubevirt.InfrastructureStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "InfrastructureStatus",
		},
	}

	return extensionscontroller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client(), infra, func() error {
		statusV1alpha1 := &apiskubevirtv1alpha1.InfrastructureStatus{}
		err := a.Scheme().Convert(status, statusV1alpha1, nil)
		if err != nil {
			return err
		}
		statusV1alpha1.SetGroupVersionKind(apiskubevirtv1alpha1.SchemeGroupVersion.WithKind("InfrastructureStatus"))
		infra.Status.ProviderStatus = &runtime.RawExtension{Object: statusV1alpha1}
		return nil
	})
}
