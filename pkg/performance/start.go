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
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

const (
	runs = 3
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) error {
	var old []float64
	var new []float64

	firstBinary := binaries[0]
	secondBinary := binaries[1]

	for r := 0; r < runs; r++ {
		log.Printf("Executing run %d...", r)
		duration, err := timeMinikubeStart(ctx, out, firstBinary)
		if err != nil {
			return errors.Wrapf(err, "timing run %d with binary %s", r, firstBinary.path)
		}
		old = append(old, duration)
		duration, err = timeMinikubeStart(ctx, out, secondBinary)
		if err != nil {
			return errors.Wrapf(err, "timing run %d with binary %s", r, secondBinary.path)
		}
		new = append(new, duration)
	}

	fmt.Fprintf(os.Stdout, " Old binary: %v\n New binary: %v\n Average Old: %f\n Average New: %f\n", old, new, average(old), average(new))

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
func timeMinikubeStart(ctx context.Context, out io.Writer, binary *Binary) (float64, error) {
	startCmd := exec.CommandContext(ctx, binary.path, startArgs(binary)...)
	startCmd.Stdout = out
	startCmd.Stderr = out

	deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
	defer deleteCmd.Run()

	log.Printf("Running: %v...", startCmd.Args)
	start := time.Now()
	if err := startCmd.Run(); err != nil {
		return 0, errors.Wrap(err, "starting minikube")
	}

	startDuration := time.Since(start).Seconds()
	return startDuration, nil
}

func startArgs(b *Binary) []string {
	args := []string{"start"}
	if b.isoURL != "" {
		args = append(args, "--iso-url", b.isoURL)
	}
	return args
}
