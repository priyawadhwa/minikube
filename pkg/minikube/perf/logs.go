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

package perf

import (
	"bufio"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func readLogs(logFile string) ([]string, error) {
	contents, err := ioutil.ReadFile(logFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s: %v", logFile, err)
	}
	return strings.Split(string(contents), "\n"), nil
}

// TimeCommandLogs runs command and watches stdout to time how long each provided log takes
func TimeCommandLogs(cmd *exec.Cmd, logsFile string) (map[string]float64, error) {
	// matches each log with the amount of time spent on that log
	var result map[string]float64

	stdout, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanBytes)

	log.Printf("Running: %v...", cmd.Args)
	if err := cmd.Start(); err != nil {
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
		result[strings.Trim(lastLog, "\n")] = timeTaken
		log.Printf("%f: %s", timeTaken, lastLog)
		lastLog = ""
	}

	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "waiting for minikube")
	}
	return result, nil
}
