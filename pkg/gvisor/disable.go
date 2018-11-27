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

package gvisor

import (
	"log"
	"path/filepath"

	"github.com/pkg/errors"
)

// Disable disables gvisor and returns state back to normal
func Disable() error {
	log.Print("Disabling gvisor...")
	// replace with old version of config.toml
	if err := rewrite(filepath.Join(nodeDir, "etc/containerd/config.toml"), defaultConfigToml); err != nil {
		return errors.Wrap(err, "rewriting config.toml")
	}
	// restart containerd
	if err := Systemctl(); err != nil {
		return errors.Wrap(err, "restarting containerd")
	}
	return nil
}
