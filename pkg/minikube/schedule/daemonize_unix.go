// +build !windows

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
	"os"
	"time"

	"github.com/VividCortex/godaemon"
)

func daemonize(profiles []string, duration time.Duration) error {
	_, _, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	if err != nil {
		return err
	}
	// now that this process has daemonized, it has a new PID
	pid := os.Getpid()
	return savePIDs(pid, profiles)
}
