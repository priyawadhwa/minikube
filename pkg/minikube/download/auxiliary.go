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

package download

import (
	"fmt"
	"runtime"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// AuxName returns name of the auxiliary tarball
func AuxName(containerRuntime string) string {
	if containerRuntime == "crio" {
		containerRuntime = "cri-o"
	}
	var storageDriver string
	if containerRuntime == "cri-o" {
		storageDriver = "overlay"
	} else {
		storageDriver = "overlay2"
	}
	return fmt.Sprintf("preloaded-images-aux-%s-%s-%s-%s.tar.lz4", AuxVersion, containerRuntime, storageDriver, runtime.GOARCH)
}

// Auxiliary downloads the auxiliary images tarball to the host
func Auxiliary(containerRuntime string) error {
	name := AuxName(containerRuntime)
	if TarballExists(name) {
		glog.Infof("Found %s in cache, skipping download", name)
		return nil
	}

	out.T(style.FileDownload, "Downloading auxiliary preload ...")
	url := remoteTarballURL(name)

	targetPath := TarballPath(name)
	if err := download(url, targetPath); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}

	if err := saveChecksumFile(name); err != nil {
		return errors.Wrap(err, "saving checksum file")
	}

	if err := verifyChecksum(name, targetPath); err != nil {
		return errors.Wrap(err, "verify")
	}
	return nil
}
