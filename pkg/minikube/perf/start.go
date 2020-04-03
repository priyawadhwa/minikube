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

package perf

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	// runs is the number of times each binary will be timed for 'minikube start'
	runs = 1
)

var (
	// For testing
	collectTimeMinikubeStart = timeMinikubeStart
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []string, logFile string) error {
	err := collectTimes(ctx, binaries, logFile)
	if err != nil {
		return err
	}

	return nil
}

func collectTimes(ctx context.Context, binaries []string, logFile string) error {

	for r := 0; r < runs; r++ {
		log.Printf("Executing run %d...", r)
		for _, binary := range binaries {
			duration, err := collectTimeMinikubeStart(ctx, binary, logFile)
			if err != nil {
				return errors.Wrapf(err, "timing run %d with %s", r, binary)
			}
			fmt.Println(duration)
		}
	}

	return nil
}

func average(nums []float64) float64 {
	total := float64(0)
	for _, a := range nums {
		total += a
	}
	return total / float64(len(nums))
}

// timeMinikubeStart returns the time it takes to execute `minikube start`
// It deletes the VM after `minikube start`.
func timeMinikubeStart(ctx context.Context, binary string, logsFile string) (map[string]float64, error) {
	startCmd := exec.CommandContext(ctx, binary, "start")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	deleteCmd := exec.CommandContext(ctx, binary, "delete")
	defer func() {
		if err := deleteCmd.Run(); err != nil {
			log.Printf("error deleting minikube: %v", err)
		}
	}()

	log.Printf("Running: %v...", startCmd.Args)
	return TimeCommandLogs(startCmd, logsFile)
}
