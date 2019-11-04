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
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	runs = 3
	// For testing
	collectTimeMinikubeStart = timeMinikubeStart
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) error {
	durations, err := collectTimes(ctx, out, binaries)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Old binary: %v\nNew binary: %v\nAverage Old: %f\nAverage New: %f\n", durations[0], durations[1], average(durations[0]), average(durations[1]))
	return nil
}

func collectTimes(ctx context.Context, out io.Writer, binaries []*Binary) ([][]float64, error) {
	durations := make([][]float64, len(binaries))
	for i := range durations {
		durations[i] = make([]float64, runs)
	}

	for r := 0; r < runs; r++ {
		log.Printf("Executing run %d/%d...", r+1, runs)
		for index, binary := range binaries {
			result, err := collectTimeMinikubeStart(ctx, out, binary)
			if err != nil {
				return nil, errors.Wrapf(err, "timing run %d with %s", r, binary.path)
			}
			durations[index][r] = result.totalTime()
		}
	}

	return durations, nil
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
func timeMinikubeStart(ctx context.Context, out io.Writer, binary *Binary) (*Result, error) {
	deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
	defer func() {
		if err := deleteCmd.Run(); err != nil {
			log.Printf("error deleting minikube: %v", err)
		}
	}()

	result := newResult()

	startCmd := exec.CommandContext(ctx, binary.path, "start")
	startCmd.Stderr = os.Stderr

	stdout, _ := startCmd.StdoutPipe()
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanBytes)

	log.Printf("Running: %v...", startCmd.Args)
	if err := startCmd.Start(); err != nil {
		return nil, err
	}

	logTimes := time.Now()

	lastLog := ""
	currentLog := ""

	for scanner.Scan() {
		text := scanner.Text()
		currentLog = currentLog + text

		if strings.Contains(currentLog, "\n") {
			lastLog = currentLog
			currentLog = ""
			continue
		}

		if !strings.Contains(lastLog, "\n") {
			continue
		}

		timeTaken := time.Since(logTimes).Seconds()
		logTimes = time.Now()
		result.addLogAndTime(lastLog, timeTaken)
		log.Printf("%f: %s", timeTaken, lastLog)
		lastLog = ""
	}

	if err := startCmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "waiting for minikube")
	}

	return result, nil
}

func startArgs(b *Binary) []string {
	args := []string{"start"}
	if b.isoURL != "" {
		args = append(args, "--iso-url", b.isoURL)
	}
	return args
}
