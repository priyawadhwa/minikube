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
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

const (
	taskName = "minikubeScheduledStop"
)

// schtasks /create /sc once /tn minikubeStop /TR "C:\Users\jenkins\minikube\out\minikube.exe stop" /ST 16:20:30 /f

func daemonize(profiles []string, duration time.Duration) error {
	currentBinary, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "getting executable")
	}
	var stopCommand string
	for _, p := range profiles {
		stopCommand = stopCommand + fmt.Sprintf("%s stop -p %s", currentBinary, p)
	}

	// the schtasks.exe command requires the stop time in HH:MM format
	stopTime := time.Now().Add(duration).Format("15:04")

	args := []string{"/CREATE", "/SC", "ONCE", "/TN", taskName, "/TR", fmt.Sprintf("\"%s\"", stopCommand), "/ST", stopTime, "/F"}
	cmd := exec.Command("schtasks.exe", args...)
	fmt.Println(cmd.Args)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))

	return nil
}
