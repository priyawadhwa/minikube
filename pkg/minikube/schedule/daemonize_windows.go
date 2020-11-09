// +build windows

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

package schedule

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/mustload"
)

func killExistingScheduledStops(profiles []string) error {
	return fmt.Errorf("not yet implemented for windows")
}

func daemonize(profiles []string, duration time.Duration) error {
	for _, profile := range profiles {
		if err := startSystemdService(profile, duration); err != nil {
			return errors.Wrapf(err, "implementing scheduled stop for %s", profile)
		}
	}
	return fmt.Errorf("not yet implemented for windows")
}

func startSystemdService(profile string, duration time.Duration) error {
	// get ssh runner
	co := mustload.Running(profile)
	command.NewSSHRunner(co.)

	// update environment file to include duration

	// restart scheduled stop service in container

	return nil
}
