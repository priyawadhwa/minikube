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

// This file implements a go-getter wrapper for cheggaaa progress bar
// based on:
// https://github.com/hashicorp/go-getter/blob/master/cmd/go-getter/progress_tracking.go

package download

import (
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/go-getter"
)

var JSONOutput getter.ProgressTracker = &jsonOutput{}

type jsonOutput struct {
	lock sync.Mutex
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *jsonOutput) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()

	fmt.Println(currentSize, totalSize)

	return &readCloser{
		Reader: stream,
		close: func() error {
			return nil
		},
	}
}
