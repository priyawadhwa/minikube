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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Result struct {
	logs  []string
	times []float64
}

func newResult() *Result {
	return &Result{
		logs:  []string{},
		times: []float64{},
	}
}

func (r *Result) addLogAndTime(log string, time float64) {
	r.logs = append(r.logs, log)
	r.times = append(r.times, time)
}

func (r *Result) totalTime() float64 {
	total := 0.0
	for _, t := range r.times {
		total += t
	}
	return total
}

func (r *Result) timeForLog(log string) (bool, float64) {
	for i, l := range r.logs {
		if l == log {
			return true, r.times[i]
		}
	}
	return false, 0
}

type DataStorage struct {
	Binaries []*Binary
	Data     map[*Binary][]*Result
}

func NewDataStorage(binaries []*Binary) *DataStorage {
	ds := &DataStorage{
		Binaries: binaries,
		Data:     map[*Binary][]*Result{},
	}
	for _, b := range binaries {
		ds.Data[b] = []*Result{}
	}
	return ds
}

func (d *DataStorage) addResult(b *Binary, result *Result) error {
	if arr, ok := d.Data[b]; ok {
		d.Data[b] = append(arr, result)
		return nil
	}
	return errors.New("unknown binary")
}

func (d *DataStorage) summarizeData(out io.Writer) error {
	for binary, results := range d.Data {
		fmt.Fprintf(out, "All Times %s: [", binary.name)
		for _, r := range results {
			fmt.Fprintf(out, " %f", r.totalTime())
		}
		fmt.Fprintf(out, "]\n")
	}
	fmt.Println()
	var averages []float64
	for binary, results := range d.Data {
		avg := averageTimeForResults(results)
		averages = append(averages, avg)
		fmt.Fprintf(out, "Average %s: **%f**\n", binary.name, avg)
	}
	fmt.Println()
	return d.summarizeTimesPerLog()
}
func (d *DataStorage) summarizeTotalTime() {

}

func (d *DataStorage) summarizeTimesPerLog() error {
	logs, err := d.logs()
	if err != nil {
		return err
	}
	binaries := d.Binaries

	table := make([][]string, len(logs))
	for i := range table {
		table[i] = make([]string, len(binaries)+1)
	}

	for i, l := range logs {
		table[i][0] = l
	}

	for i, b := range binaries {
		averageTimeForLog := d.averageTimeForLog(b)
		for log, time := range averageTimeForLog {
			index := indexForLog(logs, log)
			if index == -1 {
				continue
			}
			table[index][i+1] = fmt.Sprintf("%f", time)
		}
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Log", binaries[0].name, binaries[1].name})

	for _, v := range table {
		t.Append(v)
	}
	fmt.Println("Averages Time Per Log")
	fmt.Println("```")
	t.Render() // Send output
	fmt.Println("```")
	return nil
}

func indexForLog(logs []string, log string) int {
	for i, l := range logs {
		if strings.Contains(log, l) {
			return i
		}
	}
	return -1
}

func (d *DataStorage) logs() ([]string, error) {
	contents, err := ioutil.ReadFile("logs.txt")
	if err != nil {
		return nil, err
	}
	logs := strings.Split(string(contents), "\n")
	return logs, nil
	// return []string{"minikube v", "Creating kvm2", "Preparing Kubernetes", "Pulling images", "Launching Kubernetes", "Waiting for cluster"}
}

func (d *DataStorage) averageTimeForLog(binary *Binary) map[string]float64 {
	logToTimings := map[string][]float64{}
	for _, result := range d.Data[binary] {
		for i, l := range result.logs {
			if _, ok := logToTimings[l]; ok {
				logToTimings[l] = append(logToTimings[l], result.times[i])
				continue
			}
			logToTimings[l] = []float64{result.times[i]}
		}
	}
	logToAverageTiming := map[string]float64{}
	for log, timings := range logToTimings {
		logToAverageTiming[log] = average(timings)
	}
	return logToAverageTiming
}

func averageTimeForResults(results []*Result) float64 {
	total := 0.0
	for _, r := range results {
		total += r.totalTime()
	}
	return total / float64(len(results))
}
