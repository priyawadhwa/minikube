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
	"fmt"
	"io"
	"log"
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
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) {
	drivers := []string{"kvm2", "docker"}
	for _, d := range drivers {
		fmt.Printf("**%s Driver**\n", d)
		rm, err := collectResults(ctx, out, d, binaries)
		if err != nil {
			log.Printf("error collecting results for %s driver: %v", d, err)
			continue
		}
		rm.summarizeResults(binaries)
		fmt.Println()
	}

}

func collectResults(ctx context.Context, out io.Writer, driver string, binaries []*Binary) (*resultManager, error) {
	rm := newResultManager()
	for run := 0; run < runs; run++ {
		log.Printf("Executing run %d/%d...", run+1, runs)
		for _, binary := range binaries {
			r, err := collectTimeMinikubeStart(ctx, out, driver, binary)
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
func timeMinikubeStart(ctx context.Context, out io.Writer, driver string, binary *Binary) (*result, error) {
	downloadOnlyCmd := exec.CommandContext(ctx, binary.path, "start", fmt.Sprintf("--driver=%s", driver), "--download-only")
	if output, err := downloadOnlyCmd.CombinedOutput(); err != nil {
		log.Printf("error running download only, start may not work: %v\n%s\n", err, string(output))
	}

	startCmd := exec.CommandContext(ctx, binary.path, "start", fmt.Sprintf("--driver=%s", driver))

	deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
	defer func() {
		if err := deleteCmd.Run(); err != nil {
			log.Printf("error deleting minikube: %v", err)
		}
	}()

	log.Printf("Running: %v...", startCmd.Args)
	return timeCommandLogs(startCmd, out)
}
