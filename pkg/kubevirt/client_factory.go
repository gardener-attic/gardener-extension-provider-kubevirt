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
	"sync"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

type getClientResult struct {
	client    client.Client
	namespace string
}

// getClient retrieves and returns a client and a namespace from the given clients cache.
// If not found, it creates a client and gets a namespace using the given client factory, adds them to the cache, and returns them.
func getClient(kubeconfig []byte, clientFactory ClientFactory, clients map[string]*getClientResult, mx *sync.RWMutex) (client.Client, string, error) {
	// Check clients cache
	key := string(kubeconfig)
	mx.RLock()
	if gcr, ok := clients[key]; ok {
		mx.RUnlock()
		return gcr.client, gcr.namespace, nil
	}
	mx.RUnlock()

	// Create a new client for the given kubeconfig
	c, namespace, err := clientFactory.GetClient(kubeconfig)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create client from kubeconfig")
	}

	// Add client to cache and return it
	mx.Lock()
	clients[key] = &getClientResult{
		client:    c,
		namespace: namespace,
	}
	mx.Unlock()
	return c, namespace, nil
}
