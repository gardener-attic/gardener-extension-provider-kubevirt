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

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const KubeconfigSecretKey = "kubeconfig"

// GetKubeConfig retrieves the kubeconfig specified by the secret reference.
func GetKubeConfig(ctx context.Context, c client.Client, secretRef corev1.SecretReference) ([]byte, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, c, &secretRef)
	if err != nil {
		return []byte(""), errors.Wrap(err, "could not get secret by reference")
	}
	kubeconfig, ok := secret.Data[KubeconfigSecretKey]
	if !ok {
		return nil, errors.Errorf("missing %q field in secret", KubeconfigSecretKey)
	}
	return kubeconfig, nil
}

// GetClient creates a client from the given kubeconfig.
// It also returns the namespace of the kubeconfig's current context.
func GetClient(kubeconfig []byte) (client.Client, string, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create client config from kubeconfig")
	}
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", errors.Wrap(err, "could not get REST config from client config")
	}
	client, err := client.New(config, client.Options{})
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create client from REST config")
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", errors.Wrap(err, "could not get namespace from client config")
	}
	return client, namespace, nil
}
