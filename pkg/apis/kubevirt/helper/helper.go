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

package helper

import (
	"fmt"

	apiskubevirt "github.com/gardener/gardener-extension-provider-kubevirt/pkg/apis/kubevirt"
)

// FindMachineImage takes a list of machine images and tries to find the first entry
// whose name, version, and zone matches with the given name, version, and zone. If no such entry is
// found then an error will be returned.
func FindMachineImage(configImages []apiskubevirt.MachineImage, imageName, imageVersion string) (*apiskubevirt.MachineImage, error) {
	for _, machineImage := range configImages {
		if machineImage.Name == imageName && machineImage.Version == imageVersion {
			return &machineImage, nil
		}
	}
	return nil, fmt.Errorf("no machine image with name %q, version %q found", imageName, imageVersion)
}

// FindImage takes a list of machine images, and the desired image name and version. It tries
// to find the image with the given name and version. If it cannot be found then an error
// is returned.
func FindImage(profileImages []apiskubevirt.MachineImages, imageName, imageVersion string) (string, error) {
	for _, machineImage := range profileImages {
		if machineImage.Name == imageName {
			for _, version := range machineImage.Versions {
				if imageVersion == version.Version {
					sourceURL := ""
					if version.SourceURL != "" {
						sourceURL = version.SourceURL
					}
					return sourceURL, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find an image for name %q in version %q", imageName, imageVersion)
}
