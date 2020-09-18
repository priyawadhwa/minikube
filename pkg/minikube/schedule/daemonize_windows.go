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
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

func daemonize(profiles []string, duration time.Duration) error {
	currentBinary, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "getting current binary")
	}
	fmt.Println("Current binary is:", currentBinary)

	mkArgs := []string{"stop", "--wait", duration.String()}
	for _, p := range profiles {
		mkArgs = append(mkArgs, "-p", p)
	}
	args := append([]string{"/C", "start", fmt.Sprintf("%s", currentBinary)}, mkArgs...)
	fmt.Println("Running :", args)
	cmd := exec.Command("cmd.exe", args...)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "starting daemon command")
	}
	fmt.Println("Need to store process: ", cmd.Process.Pid)
	return savePIDs(cmd.Process.Pid, profiles)
}
