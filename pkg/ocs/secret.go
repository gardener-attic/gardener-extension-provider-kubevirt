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

package ocs

import (
	"context"
	"fmt"
	"strconv"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AccessKeyID is a constant for the key in a backup secret that holds the OCS S3 access key id.
	AccessKeyID = "accessKeyID"
	// SecretAccessKey is a constant for the key in a backup secret that holds the OCS S3 secret access key.
	SecretAccessKey = "secretAccessKey"
	// Endpoint is a constant for the key in a backup secret that holds an OCS S3 endpoint.
	Endpoint = "endpoint"
	// DisableSSL is a constant for the key in a backup secret that specifies whether SSL should be disabled or not.
	DisableSSL = "disableSSL"
	// InsecureSkipVerify is a constant for the key in a backup secret that specifies whether the client verifies the server's certificate chain and host name.
	InsecureSkipVerify = "insecureSkipVerify"
	// Region is a constant for the key in a backup secret that points to a region.
	Region = "region"
)

// Credentials stores AWS credentials.
type Credentials struct {
	AccessKeyID        string
	SecretAccessKey    string
	Endpoint           string
	DisableSSL         bool
	InsecureSkipVerify bool
}

// GetCredentialsFromSecretRef reads the secret given by the the secret reference and returns the read Credentials
// object.
func GetCredentialsFromSecretRef(ctx context.Context, client client.Client, secretRef corev1.SecretReference) (*Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, client, &secretRef)
	if err != nil {
		return nil, err
	}
	return ReadCredentialsSecret(secret)
}

// ReadCredentialsSecret reads a secret containing credentials.
func ReadCredentialsSecret(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret does not contain any data")
	}

	accessKeyID, ok := secret.Data[AccessKeyID]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", AccessKeyID)
	}
	secretAccessKey, ok := secret.Data[SecretAccessKey]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", SecretAccessKey)
	}
	endpoint, ok := secret.Data[Endpoint]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", Endpoint)
	}
	disableSSL, err := getBool(secret, DisableSSL)
	if err != nil {
		return nil, err
	}
	insecureSkipVerify, err := getBool(secret, InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		AccessKeyID:        string(accessKeyID),
		SecretAccessKey:    string(secretAccessKey),
		Endpoint:           string(endpoint),
		DisableSSL:         disableSSL,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

// NewClientFromSecretRef creates a new s3 Client for the given s3 credentials from given k8s <secretRef> and
// the kubevirt <region>.
func NewClientFromSecretRef(ctx context.Context, client client.Client, secretRef corev1.SecretReference, region string) (*Client, error) {
	credentials, err := GetCredentialsFromSecretRef(ctx, client, secretRef)
	if err != nil {
		return nil, err
	}
	return NewClient(credentials.AccessKeyID, credentials.SecretAccessKey, credentials.Endpoint, region, credentials.DisableSSL, credentials.InsecureSkipVerify)
}

func getBool(secret *corev1.Secret, key string) (bool, error) {
	result := false
	value := secret.Data[key]
	if len(value) > 0 {
		var err error
		if result, err = strconv.ParseBool(string(value)); err != nil {
			return false, errors.Wrapf(err, "could not parse %q field in secret to bool", key)
		}
	}
	return result, nil
}
