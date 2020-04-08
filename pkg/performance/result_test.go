/*
Copyright 2017 The Kubernetes Authors All rights reserved.
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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAverageTimePerLog(t *testing.T) {
	results := []*result{
		{
			timedLogs: map[string]float64{
				"first log":  10,
				"second log": 20,
			},
		}, {
			timedLogs: map[string]float64{
				"first log": 12,
				"third log": 4,
			},
		},
	}

	expected := map[string]float64{
		"first log":  11.0,
		"second log": 20.0,
		"third log":  4.0,
	}

	actual := averageTimePerLog(results)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("results mismatch (-want +got):\n%s", diff)
	}
}
