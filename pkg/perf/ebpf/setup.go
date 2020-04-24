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
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	url         = "https://storage.googleapis.com/minikube-kernel-headers/kernel-headers-linux-4.19.94.tar.lz4"
	tarballDest = "/tmp/kernel-headers-linux-4.19.94.tar.lz4"
	dest        = "/node/lib/modules/4.19.94/build"
)

func Setup() error {
	// else, download kernel modules to tmpDest
	if err := downloadKernelModules(); err != nil {
		return errors.Wrap(err, "downloading kernel modules")
	}
	if err := createDir(); err != nil {
		return errors.Wrap(err, "creating dir")
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
		Src:  url,
		Dst:  tmpDst,
		Mode: getter.ClientModeFile,
	}
	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}
	fmt.Println("renaming", tmpDst, tarballDest)
	return os.Rename(tmpDst, tarballDest)
}

func createDir() error {
	fmt.Println("Creating", dest)
	cmd := exec.Command("mkdir", "-p", dest)
	fmt.Println(cmd.Args)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "creating dir: %v", output)
	}
	return nil
}

func extractKernelModules() error {
	fmt.Println("Exctracting kernel modules...")
	cmd := exec.Command("tar", "-I", "lz4", "-C", dest, "-xvf", tarballDest)
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	fmt.Println(cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "extracting kernel modules: %s", stderr.String())
	}
	return nil
}

func removeTarball() error {
	cmd := exec.Command("rm", "-rf", tarballDest)
	fmt.Println(cmd.Args)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "removing tarball: %v", output)
	}
	return nil
}
