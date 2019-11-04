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
