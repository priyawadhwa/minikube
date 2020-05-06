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

const (
	PreparingKubernetes = "preparing_kubernetes"
)

type LogType string

// All the Style constants available
const (
	Log     LogType = "Log"
	Warning LogType = "Warning"
	Error   LogType = "Error"
)

type log struct {
	LogType
	style   StyleEnum
	message string
	v       []V
}

// Registry holds all user-facing logs
type Registry struct {
	Logs  map[string]log
	Index int
}

var registry Registry

// Init initializes the logs registry
func Init() {
	registry = Registry{}
}

// Register registers a log
func Register(name string, style StyleEnum, message string, logType LogType) {
	registry.Logs[name] = log{
		style:   style,
		message: message,
		LogType: logType,
	}
}
