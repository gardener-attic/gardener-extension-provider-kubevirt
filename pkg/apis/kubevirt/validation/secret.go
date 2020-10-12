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

package validation

import (
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// ValidateCloudProviderSecret checks whether the given secret contains a valid kubeconfig.
func ValidateCloudProviderSecret(secret *corev1.Secret) error {
	kubeconfig, ok := secret.Data[kubevirt.Kubeconfig]
	if !ok {
		return errors.Errorf("missing %q field in secret", kubevirt.Kubeconfig)
	}

	if _, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig); err != nil {
		return errors.Wrapf(err, "could not create REST config from kubeconfig")
	}

	return nil
}
