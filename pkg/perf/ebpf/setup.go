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

package ebpf

import (
	"os"
	"os/exec"

	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	url         = "https://storage.googleapis.com/minikube-kernel-headers/kernel-headers-linux-4.19.94.tar.lz4"
	tarballDest = "/tmp/kernel-headers-linux-4.19.94.tar.lz4"
	dest        = "/lib/modules/4.19.94/build"
)

func setup() error {
	if fi, err := os.Stat(dest); err == nil && fi.IsDir() {
		glog.Infof("Kernel modules have already been downloaded, skipping...")
		return nil
	}
	// else, download kernel modules to tmpDest
	if err := downloadKernelModules(); err != nil {
		return errors.Wrap(err, "downloading kernel modules")
	}
	// extract kernel modules to dest
	if err := extractKernelModules(); err != nil {
		return errors.Wrap(err, "extracting kernel modules")
	}
	// delete downloaded tarball
	return removeTarball()
}

func downloadKernelModules() error {
	out.T(out.FileDownload, "Downloading kernel modules ...")

	tmpDst := tarballDest + ".download"
	client := &getter.Client{
		Src:     url,
		Dst:     tmpDst,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(download.DefaultProgressBar)},
	}
	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}
	return os.Rename(tmpDst, tarballDest)
}

func createDir() error {
	cmd := exec.Command("sudo", "mkdir", "-p", dest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "creating dir: %v", output)
	}
	return nil
}

func extractKernelModules() error {
	cmd := exec.Command("sudo", "tar", "-I", "lz4", "-C", dest, "-xvf", tarballDest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "extracting kernel modules: %v", output)
	}
	return nil
}

func removeTarball() error {
	cmd := exec.Command("sudo", "rm", "-rf", tarballDest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "removing tarball: %v", output)
	}
	return nil
}
