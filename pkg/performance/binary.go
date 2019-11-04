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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

type Binary struct {
	name   string
	path   string
	isoURL string
	pr     int
}

const (
	prPrefix = "pr://"
)

func NewBinary(b string) (*Binary, error) {
	// If it doesn't have the prefix, assume a path
	if !strings.HasPrefix(b, prPrefix) {
		return &Binary{
			path: b,
		}, nil
	}
	return newBinaryFromPR(b)
}

func (b *Binary) SetName(name string) {
	b.name = name
}

func newBinaryFromPR(pr string) (*Binary, error) {
	pr = strings.TrimPrefix(pr, prPrefix)
	// try to convert to int
	i, err := strconv.Atoi(pr)
	if err != nil {
		return nil, errors.Wrapf(err, "converting %s to an integer", pr)
	}

	b := &Binary{
		path:   localMinikubePath(i),
		isoURL: remoteMinikubeIsoURL(i),
		pr:     i,
		name:   fmt.Sprintf("Minikube (PR %d)", i),
	}

	if err := downloadBinary(remoteMinikubeURL(i), b.path); err != nil {
		return nil, errors.Wrapf(err, "downloading minikube")
	}

	return b, nil
}

func remoteMinikubeURL(pr int) string {
	return fmt.Sprintf("https://storage.googleapis.com/minikube-builds/%d/minikube-linux-amd64", pr)
}

// remoteMinikubeIsoURL returns the url for minikube.iso if it exists for this PR
func remoteMinikubeIsoURL(pr int) string {
	url := fmt.Sprintf("https://storage.googleapis.com/minikube-builds/%d/minikube.iso", pr)
	// Return url if it exists
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	return url
}

func localMinikubePath(pr int) string {
	return fmt.Sprintf("%s/minikube-binaries/%d/minikube", constants.DefaultMinipath, pr)
}

func downloadBinary(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
