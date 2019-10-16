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

package performance

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

const (
	runs = 1
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, firstBinary, secondBinary string) error {
	var old []float64
	var new []float64

	for r := 0; r < runs; r++ {
		log.Printf("Executing run %d...", r)
		duration, err := timeMinikubeStart(ctx, firstBinary)
		if err != nil {
			return errors.Wrapf(err, "timing run %d with binary %s", r, firstBinary)
		}
		old = append(old, duration)
		duration, err = timeMinikubeStart(ctx, secondBinary)
		if err != nil {
			return errors.Wrapf(err, "timing run %d with binary %s", r, secondBinary)
		}
		new = append(new, duration)
	}

	fmt.Printf(" Old binary: %v\n New binary: %v\n Average Old: %f\n Average New: %f\n", old, new, average(old), average(new))

	return nil
}

func average(array []float64) float64 {
	total := float64(0)
	for _, a := range array {
		total += a
	}
	return total / float64(len(array))
}

// timeMinikubeStart returns the time it takes to execute `minikube start`
// It deletes the VM after `minikube start`.
func timeMinikubeStart(ctx context.Context, binary string) (float64, error) {
	startCmd := exec.CommandContext(ctx, binary, "start")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	deleteCmd := exec.CommandContext(ctx, binary, "delete")
	defer deleteCmd.Run()

	log.Printf("Running `minikube start` with %s...", binary)
	start := time.Now()
	if err := startCmd.Run(); err != nil {
		return 0, errors.Wrap(err, "starting minikube")
	}

	startDuration := time.Since(start).Seconds()
	return startDuration, nil
}
