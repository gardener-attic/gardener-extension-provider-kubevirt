// Copyright (c) 2020 SAP SE or an SAP affiliate company and Gardener Project Authors. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const (
	// ErrCodeNoSuchBucket for service response error code "BucketNotEmpty". The bucket you tried to delete is not empty.
	ErrCodeBucketNotEmpty = "BucketNotEmpty"
)

type Client struct {
	S3 *s3.S3
}

func NewClient(accessKey, secret, endpoint, region string) (*Client, error) {
	noSSL := false
	pair := strings.Split(endpoint, "://")
	if len(pair) > 1 {
		noSSL = pair[0] == "http"
	}

	pathStyle := true
	token := ""
	creds := credentials.NewStaticCredentials(accessKey, secret, token)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:           &region,
			Credentials:      creds,
			Endpoint:         &endpoint,
			DisableSSL:       &noSSL,
			S3ForcePathStyle: &pathStyle,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create S3 session instance")
	}

	return &Client{
		S3: s3.New(sess),
	}, nil
}

// CreateBucketIfNotExists creates the s3 bucket with name <bucket> in <region>.
func (c *Client) CreateBucketIfNotExists(ctx context.Context, bucket, region string) error {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(s3.BucketCannedACLPrivate),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	}
	if _, err := c.S3.CreateBucketWithContext(ctx, createBucketInput); err != nil {
		if aerr, ok := err.(awserr.Error); !ok {
			return errors.Wrap(err, "could not create bucket")
		} else if aerr.Code() != s3.ErrCodeBucketAlreadyExists && aerr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
			return errors.Wrap(err, "could not create bucket")
		}
	}

	// Set lifecycle rule to purge incomplete multipart upload orphaned because of force shutdown or rescheduling or networking issue with etcd-backup-restore.
	putBucketLifecycleConfigurationInput := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: []*s3.LifecycleRule{
				{
					// Note: Though as per documentation at https://docs.aws.amazon.com/AmazonS3/latest/API/API_LifecycleRule.html the Filter field is
					// optional, if not specified the SDK API fails with `Malformed XML` error code. Cross verified same behavior with aws-cli client as well.
					// Please do not remove it.
					Filter: &s3.LifecycleRuleFilter{
						Prefix: aws.String(""),
					},
					AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
						DaysAfterInitiation: aws.Int64(7),
					},
					Status: aws.String(s3.ExpirationStatusEnabled),
				},
			},
		},
	}

	_, err := c.S3.PutBucketLifecycleConfigurationWithContext(ctx, putBucketLifecycleConfigurationInput)
	return errors.Wrap(err, "could not put bucket lifecycle configuration")
}

// DeleteBucketIfExists deletes the s3 bucket with name <bucket>. If it does not exist,
// no error is returned.
func (c *Client) DeleteBucketIfExists(ctx context.Context, bucket string) error {
	if _, err := c.S3.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchBucket {
				return nil
			}
			if aerr.Code() == ErrCodeBucketNotEmpty {
				if err := c.DeleteObjectsWithPrefix(ctx, bucket, ""); err != nil {
					return err
				}
				return c.DeleteBucketIfExists(ctx, bucket)
			}
		}
		return errors.Wrap(err, "could not delete bucket")
	}
	return nil
}

// DeleteObjectsWithPrefix deletes the s3 objects with the specific <prefix> from <bucket>. If it does not exist,
// no error is returned.
func (c *Client) DeleteObjectsWithPrefix(ctx context.Context, bucket, prefix string) error {
	in := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	var delErr error
	if err := c.S3.ListObjectsPagesWithContext(ctx, in, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		objectIDs := make([]*s3.ObjectIdentifier, 0)
		for _, key := range page.Contents {
			obj := &s3.ObjectIdentifier{
				Key: key.Key,
			}
			objectIDs = append(objectIDs, obj)
		}

		if len(objectIDs) != 0 {
			if _, delErr = c.S3.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket),
				Delete: &s3.Delete{
					Objects: objectIDs,
					Quiet:   aws.Bool(true),
				},
			}); delErr != nil {
				return false
			}
		}
		return !lastPage
	}); err != nil {
		return errors.Wrap(err, "could not list objects")
	}
	if delErr != nil {
		if aerr, ok := delErr.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil
		}
		return errors.Wrap(delErr, "could not delete objects")
	}
	return nil
}
