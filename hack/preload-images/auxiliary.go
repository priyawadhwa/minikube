/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
)

// create auxiliary tarball if needed
func auxiliary(cr string) error {
	aux := download.AuxName(cr)
	if download.TarballExists(aux) {
		fmt.Printf("Auxiliary tarball for version %v and runtime %s already exists, skipping generation\n", download.AuxVersion, cr)
		return nil
	}
	imgs := images.Auxiliary("")
	if err := generateTarball(imgs, constants.DefaultKubernetesVersion, cr, aux); err != nil {
		exit(fmt.Sprintf("generating tarball"), err)
	}
	if err := uploadTarball(aux); err != nil {
		exit(fmt.Sprintf("uploading tarball"), err)
	}

	if err := deleteMinikube(); err != nil {
		fmt.Printf("error cleaning up minikube before finishing up: %v\n", err)
	}

	return nil
}
