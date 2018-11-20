/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package gvisor

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	nodeDir = "/node"
)

// Enable follows these steps for enabling gvisor in minikube:
//   1. rewrites the /etc/conntainerd/config.toml on the host  (or is it /run/docker/containerd/containerd.toml?)
//   2. downloads gvisor + shim
//   3. restarts containerd (TODO: see if this actually works) with docker run --pid="host" -v /bin:/bin   -v /etc:/etc -v /usr/lib:/usr/lib -v /usr/share:/usr/share -v /tmp:/tmp -v /run/systemd:/run/systemd -v /sys:/sys -v /mnt:/mnt -v /var/lib:/var/lib -v /usr/libexec:/usr/libexec gcr.io/priya-wadhwa/gvisor:latest
func Enable() error {
	if err := makeDirs(); err != nil {
		return errors.Wrap(err, "creating directories on node")
	}
	if err := downloadBinaries(); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}
	if err := copyFiles(); err != nil {
		return errors.Wrap(err, "copying files")
	}
	if err := systemctl(); err != nil {
		return errors.Wrap(err, "systemctl")
	}
	return nil
}

func makeDirs() error {
	// Make /run/containerd/runsc to hold logs
	fp := filepath.Join(nodeDir, "run/containerd/runsc")
	if err := os.MkdirAll(fp, 0755); err != nil {
		return errors.Wrap(err, "creating runsc dir")
	}

	// Make /usr/local/bin to store the runsc binary
	fp = filepath.Join(nodeDir, "usr/local/bin")
	if err := os.MkdirAll(fp, 0755); err != nil {
		return errors.Wrap(err, "creating usr/local/bin dir")
	}

	fp = filepath.Join(nodeDir, "tmp/runsc")
	if err := os.MkdirAll(fp, 0755); err != nil {
		return errors.Wrap(err, "creating runsc logs dir")
	}

	return nil
}

func downloadBinaries() error {
	if err := runsc(); err != nil {
		return errors.Wrap(err, "downloading runsc")
	}
	if err := gvisorContainerdShim(); err != nil {
		return errors.Wrap(err, "downloading gvisor-containerd-shim")
	}
	return nil
}

// downloads the gvisor-containerd-shim
func gvisorContainerdShim() error {
	dest := filepath.Join(nodeDir, "usr/bin/gvisor-containerd-shim")
	return wget("http://storage.googleapis.com/balintp-minikube/gvisor-containerd-shim", dest)
}

// downloads the runsc binary and returns a path to the binary
func runsc() error {
	dest := filepath.Join(nodeDir, "usr/local/bin/runsc")
	return wget("http://storage.googleapis.com/gvisor/releases/nightly/latest/runsc", dest)
}

func wget(url, dest string) error {
	cmd := exec.Command("wget", url)
	cmd.Dir = filepath.Dir(dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "downloading binary")
	}
	if err := os.Chmod(dest, 0777); err != nil {
		return errors.Wrap(err, "fixing perms")
	}
	return nil
}

func curlFile(url, dest string) error {
	cmd := exec.Command("curl", "-X", "GET", "-L0", url, "--output", filepath.Base(dest))
	cmd.Dir = filepath.Dir(dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrapf(err, "downloading %s", filepath.Base(dest))
	}
	return nil
}

func downloadFileToDest(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fi, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "creating %s", dest)
	}
	defer fi.Close()
	if _, err := io.Copy(fi, resp.Body); err != nil {
		return errors.Wrap(err, "copying binary")
	}
	if err := fi.Chmod(0777); err != nil {
		return errors.Wrap(err, "fixing perms")
	}
	return nil
}

// Must rewrite the following files:
//    1. gvisor-containerd-shim.toml
//    2. containerd config.toml
func copyFiles() error {
	if err := rewriteContainerdToml(); err != nil {
		return errors.Wrap(err, "rewriting config.toml")
	}
	if err := rewriteShimToml(); err != nil {
		return errors.Wrap(err, "rewriting gvisor-containerd-shim.toml")
	}
	return nil
}

func rewriteShimToml() error {
	// delete the current shim.toml and replace it with the one we want
	path := filepath.Join(nodeDir, "etc/containerd/gvisor-containerd-shim.toml")
	// Now, create the new shim.toml
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "creating %s", path)
	}
	if _, err := f.Write([]byte(gvisorShim)); err != nil {
		return errors.Wrap(err, "writing gvisor-containerd-shim.toml")
	}
	return nil
}

func rewriteContainerdToml() error {
	// delete the current config.toml and replace it with the one we want
	path := filepath.Join(nodeDir, "etc/containerd/config.toml")
	if err := os.Remove(path); err != nil {
		return errors.Wrap(err, "removing config.toml")
	}
	// Now, create the new config.toml
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "creating %s", path)
	}
	if _, err := f.Write([]byte(configToml)); err != nil {
		return errors.Wrap(err, "writing config.toml")
	}
	return nil
}

func systemctl() error {
	dir := filepath.Join(nodeDir, "usr/libexec/sudo")
	if err := os.Setenv("LD_LIBRARY_PATH", dir); err != nil {
		return errors.Wrap(err, dir)
	}

	log.Print("Trying to stop rpc-statd.service")
	// first, stop  rpc-statd.service
	cmd := exec.Command("sudo", "-E", "systemctl", "stop", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return errors.Wrap(err, "stopping rpc-statd.service")
	}
	// restart containerd
	cmd = exec.Command("sudo", "-E", "systemctl", "restart", "containerd")
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "restarting containerd")
	}
	// start rpd-statd.service
	cmd = exec.Command("sudo", "-E", "systemctl", "start", "rpc-statd.service")
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "restarting rpc-statd.service")
	}
	return nil
}
