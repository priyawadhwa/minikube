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
	"time"

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
	if err := Systemctl(); err != nil {
		return errors.Wrap(err, "restarting containerd")
	}
	// sleep for one year so the pod continuously runs
	time.Sleep(24 * 7 * 52 * time.Hour)
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
	return downloadFileToDest("http://storage.googleapis.com/balintp-minikube/gvisor-containerd-shim", dest)
}

// downloads the runsc binary and returns a path to the binary
func runsc() error {
	dest := filepath.Join(nodeDir, "usr/local/bin/runsc")
	return downloadFileToDest("http://storage.googleapis.com/gvisor/releases/nightly/latest/runsc", dest)
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

func downloadFileToDest(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := os.Stat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return errors.Wrapf(err, "removing %s for overwrite", dest)
		}
	}
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
	if err := rewrite(filepath.Join(nodeDir, "etc/containerd/config.toml"), gvisorConfigToml); err != nil {
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

func rewrite(path, contents string) error {
	if err := os.Remove(path); err != nil {
		return errors.Wrapf(err, "removing %s", path)
	}
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "creating %s", path)
	}
	if _, err := f.Write([]byte(contents)); err != nil {
		return errors.Wrap(err, "writing config.toml")
	}
	return nil
}

//   3. restarts containerd (TODO: see if this actually works) with docker run --pid="host" -v /bin:/bin   -v /etc:/etc -v /usr/lib:/usr/lib -v /usr/share:/usr/share -v /tmp:/tmp -v /run/systemd:/run/systemd -v /sys:/sys -v /mnt:/mnt -v /var/lib:/var/lib -v /usr/libexec:/usr/libexec gcr.io/priya-wadhwa/gvisor:latest

// Restart runs a docker container which will restart containerd
func Restart() error {
	log.Print("Trying to restart containerd...")
	mounts := []string{
		"/bin", // this will mount in systemctl
		// the following mount in libraries needed by systemctl
		"/usr/lib/systemd",
		"/mnt",
		"/usr/libexec/sudo",
		"/var/lib",
		"/etc",
		"/run/systemd",
		"/usr/bin",
	}
	args := []string{"run", "--pid=host"}
	for _, m := range mounts {
		args = append(args, []string{"-v", fmt.Sprintf("%s:%s", m, m)}...)
	}
	args = append(args, []string{image, "/gvisor", "-restart"}...)

	cmd := exec.Command("docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "running docker command")
	}
	return nil
}

func Systemctl() error {
	dir := filepath.Join(nodeDir, "usr/libexec/sudo")
	if err := os.Setenv("LD_LIBRARY_PATH", dir); err != nil {
		return errors.Wrap(err, dir)
	}

	log.Print("Stopping rpc-statd.service...")
	// first, stop  rpc-statd.service
	cmd := exec.Command("sudo", "-E", "systemctl", "stop", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return errors.Wrap(err, "stopping rpc-statd.service")
	}
	// restart containerd
	log.Print("Restarting containerd...")
	cmd = exec.Command("sudo", "-E", "systemctl", "restart", "containerd")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "restarting containerd")
	}
	// start rpc-statd.service
	log.Print("Starting rpc-statd...")
	cmd = exec.Command("sudo", "-E", "systemctl", "start", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "restarting rpc-statd.service")
	}
	log.Print("containerd restart complete")
	return nil
}
