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

// Package out provides a mechanism for sending localized, stylized output to the console.
package out

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
)

const (
	MinikubeVersion      = "Minikube Version"
	SelectDriver         = "Selecting Driver"
	StartingControlPlane = "Starting Control Plane"
	DownloadArtifacts    = "Download Necessary Artifacts"
	CreatingNode         = "Creating Node"
	PreparingKubernetes  = "Preparing Kubernetes"
	VerifyingKubernetes  = "Verifying Kubernetes"
	EnablingAddons       = "Enabling Addons"
	Done                 = "Done"
)

type LogType string

// All the Style constants available
const (
	Log         LogType = "Log"
	WarningType LogType = "Warning"
	Error       LogType = "Error"
)

type log struct {
	Name        string
	Message     string
	TotalSteps  int
	CurrentStep int
}

// Registry holds all user-facing logs
type Registry struct {
	Logs  map[string]*log
	Index int
}

var registry Registry

// Init initializes the logs registry
func Init() {
	registry = Registry{
		Logs:  map[string]*log{},
		Index: 1,
	}
	Register(MinikubeVersion)
	Register(SelectDriver)
	Register(StartingControlPlane)
	Register(DownloadArtifacts)
	Register(CreatingNode)
	Register(PreparingKubernetes)
	Register(VerifyingKubernetes)
	if viper.GetBool("install-addons") {
		Register(EnablingAddons)
	}
	Register(Done)
}

// Register registers a log
func Register(name string) {
	registry.Logs[name] = &log{
		Name:        name,
		CurrentStep: registry.index(),
	}
	registry.increaseIndex()
}

func (r *Registry) NumLogs() int {
	return len(r.Logs)
}

func (r *Registry) increaseIndex() {
	r.Index++
}

func (r *Registry) index() int {
	return r.Index
}

func updateLog(name string, message string) (*log, error) {
	l := registry.Logs[name]
	if l == nil {
		return nil, fmt.Errorf("no log called %s exists in registry", name)
	}
	l.Message = message
	l.TotalSteps = registry.NumLogs()
	return l, nil
}

func JsonEncoding(name string, message string) (string, error) {
	log, err := updateLog(name, message)
	if err != nil {
		return "", err
	}
	encoding, err := json.Marshal(log)
	return string(encoding), err
}
