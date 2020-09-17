// +build integration

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

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/util/retry"
)

func TestScheduledStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("feature not yet implemented for windows")
	}
	profile := UniqueProfileName("scheduled-stop")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)
	startMinikube(ctx, t, profile)

	// schedule a stop for 5 min from now and make sure PID is created
	scheduledStopMinikube(ctx, t, profile, "5m")
	pid := checkPID(t, profile)
	if !processRunning(t, pid) {
		t.Fatalf("process %v is not running", pid)
	}

	// redo scheduled stop to be 3 min
	scheduledStopMinikube(ctx, t, profile, "10s")
	if processRunning(t, pid) {
		t.Fatalf("process %v running but should have been killed on reschedule of stop", pid)
	}
	checkPID(t, profile)
	// wait allotted time to make sure minikube status is "Stopped"
	time.Sleep(15 * time.Second)
	checkStatus := func() error {
		got := Status(ctx, t, Target(), profile, "Host", profile)
		if got != state.Stopped.String() {
			return fmt.Errorf("expected post-stop host status to be -%q- but got *%q*", state.Stopped, got)
		}
		return nil
	}
	if err := retry.Expo(checkStatus, 100*time.Microsecond, 5*time.Second); err != nil {
		t.Fatalf("error %v", err)
	}
}

func startMinikube(ctx context.Context, t *testing.T, profile string) {
	args := append([]string{"start", "-p", profile}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}
}

func scheduledStopMinikube(ctx context.Context, t *testing.T, profile string, stop string) {
	args := []string{"stop", "-p", profile, "--schedule", stop}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}
}

func checkPID(t *testing.T, profile string) string {
	file := localpath.PID(profile)
	var contents []byte
	getContents := func() error {
		var err error
		contents, err = ioutil.ReadFile(file)
		return err
	}
	// first, make sure the PID file exists
	if err := retry.Expo(getContents, 100*time.Microsecond, time.Minute*1); err != nil {
		t.Fatalf("error reading %s: %v", file, err)
	}
	return string(contents)
}

func processRunning(t *testing.T, pid string) bool {
	// make sure PID file contains a running process
	p, err := strconv.Atoi(pid)
	if err != nil {
		return false
	}
	process, err := os.FindProcess(p)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	t.Log("signal error was: ", err)
	return err == nil
}
