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

package performance

import (
	"context"
	"io"
	"log"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	// runs is the number of times each binary will be timed for 'minikube start'
	runs = 3
)

var (
	// For testing
	collectTimeMinikubeStart = timeMinikubeStart
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) error {
	rm, err := collectResults(ctx, binaries)
	if err != nil {
		return err
	}
	rm.summarizeResults(binaries)
	return nil
}

func collectResults(ctx context.Context, binaries []*Binary) (*resultManager, error) {
	rm := newResultManager()
	for run := 0; run < runs; run++ {
		log.Printf("Executing run %d/%d...", run+1, runs)
		for _, binary := range binaries {
			r, err := collectTimeMinikubeStart(ctx, binary)
			if err != nil {
				return nil, errors.Wrapf(err, "timing run %d with %s", run, binary)
			}
			rm.addResult(binary, r)
		}
	}
	return rm, nil
}

// timeMinikubeStart returns the time it takes to execute `minikube start`
// It deletes the VM after `minikube start`.
func timeMinikubeStart(ctx context.Context, binary *Binary) (*result, error) {
	startCmd := exec.CommandContext(ctx, binary.path, "start")

	deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
	defer func() {
		if err := deleteCmd.Run(); err != nil {
			log.Printf("error deleting minikube: %v", err)
		}
	}()

	log.Printf("Running: %v...", startCmd.Args)
	return timeCommandLogs(startCmd)
}
