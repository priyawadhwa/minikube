/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package docker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	priority := registry.Default
	// Staged rollout for preferred:
	// - Linux
	// - Windows (once "service" command works)
	// - macOS
	if runtime.GOOS == "linux" {
		priority = registry.Preferred
	}

	if err := registry.Register(registry.DriverDef{
		Name:     driver.Docker,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIBinary: oci.Docker}) },
		Status:   status,
		Priority: priority,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	return kic.NewDriver(kic.Config{
		MachineName:       driver.MachineName(cc, n),
		StorePath:         localpath.MiniPath(),
		ImageDigest:       kic.BaseImage,
		CPU:               cc.CPUs,
		Memory:            cc.Memory,
		OCIBinary:         oci.Docker,
		APIServerPort:     cc.Nodes[0].Port,
		KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		ContainerRuntime:  cc.KubernetesConfig.ContainerRuntime,
	}), nil
}

func status() registry.State {
	_, err := exec.LookPath(oci.Docker)
	if err != nil {
		return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Docker is required.", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/docker/"}
	}

	// Allow no more than 3 seconds for docker info
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = exec.CommandContext(ctx, oci.Docker, "info").Run()
	if err != nil {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Docker is not running or is responding too slow. Try: restarting docker desktop."}
	}

	return registry.State{Installed: true, Healthy: true}
}
