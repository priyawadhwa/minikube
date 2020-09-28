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
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows/svc/mgr"
)

func daemonize(profiles []string, duration time.Duration) error {
	currentBinary, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "getting executable")
	}
	fmt.Println(currentBinary)
	// cmd := exec.Command(currentBinary, "stop", "--wait", duration.String())
	// if err := cmd.Start(); err != nil {
	// 	fmt.Println(err)
	// }
	// return savePIDs(cmd.Process.Pid, profiles)

	m, err := mgr.Connect()
	if err != nil {
		return errors.Wrap(err, "getting manager")
	}

	svcName := "minikubeScheduleStop"
	svc, err := m.CreateService(svcName, currentBinary, mgr.Config{}, []string{"--wait", fmt.Sprintf("%v", duration.Seconds())}...)

	if err != nil {
		return errors.Wrap(err, "creating service")
	}
	if err := svc.Start(); err != nil {
		return errors.Wrap(err, "starting service")
	}
	// TODO: get PID of the service
	// save it via return savePIDs(<pid>, profiles)
	return nil
}
