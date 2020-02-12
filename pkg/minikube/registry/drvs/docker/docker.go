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
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.Docker,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIBinary: oci.Docker}) },
		Status:   status,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(mc config.MachineConfig) interface{} {
	baseImage, _ := GetKicBaseimage(mc.KubernetesConfig.KubernetesVersion)
	return kic.NewDriver(kic.Config{
		MachineName:   mc.Name,
		StorePath:     localpath.MiniPath(),
		ImageDigest:   baseImage,
		CPU:           mc.CPUs,
		Memory:        mc.Memory,
		OCIBinary:     oci.Docker,
		APIServerPort: mc.Nodes[0].Port,
	})

}

// GetKicBaseimage returns a base image with k8s images preloaded if available
// otherwise, it returns the default base image
func GetKicBaseimage(kv string) (string, bool) {
	image := fmt.Sprintf("gcr.io/k8s-minikube/kicbase:v0.0.5-k8s-%s", kv)
	// Check if image exists locally or remotely
	ref, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return kic.BaseImage, false
	}
	_, err = remote.Image(ref)
	if err == nil {
		return image, true
	}
	return kic.BaseImage, false
}

func status() registry.State {
	_, err := exec.LookPath("docker")
	if err != nil {
		return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Docker is required.", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/kic/"}
	}
	// Allow no more than 2 seconds for querying state
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = exec.CommandContext(ctx, "docker", "info").Run()
	if err != nil {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Docker is not running. Try: restarting docker desktop."}
	}

	return registry.State{Installed: true, Healthy: true}
}
