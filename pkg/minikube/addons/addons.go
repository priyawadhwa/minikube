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

package addons

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

// EnableOrDisableAddon updates addon status executing any commands necessary
func EnableOrDisableAddon(name string, val string, profile string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	addon := assets.Addons[name]

	// check addon status before enabling/disabling it
	alreadySet, err := isAddonAlreadySet(addon, enable)
	if err != nil {
		out.ErrT(out.Conflict, "{{.error}}", out.V{"error": err})
		return err
	}
	//if addon is already enabled or disabled, do nothing
	if alreadySet {
		return nil
	}

	// TODO(r2d4): config package should not reference API, pull this out
	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "machine client")
	}
	defer api.Close()

	//if minikube is not running, we return and simply update the value in the addon
	//config and rewrite the file
	if !cluster.IsMinikubeRunning(api) {
		return nil
	}

	cfg, err := config.Load(viper.GetString(config.MachineProfile))
	if err != nil && !os.IsNotExist(err) {
		exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	host, err := cluster.CheckIfHostExistsAndLoad(api, cfg.Name)
	if err != nil {
		return errors.Wrap(err, "getting host")
	}

	cmd, err := machine.CommandRunner(host)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	data := assets.GenerateTemplateData(cfg.KubernetesConfig)
	return enableOrDisableAddonInternal(addon, cmd, data, enable)
}

func isAddonAlreadySet(addon *assets.Addon, enable bool) (bool, error) {
	addonStatus, err := addon.IsEnabled()

	if err != nil {
		return false, errors.Wrap(err, "get the addon status")
	}

	if addonStatus && enable {
		return true, nil
	} else if !addonStatus && !enable {
		return true, nil
	}

	return false, nil
}

func enableOrDisableAddonInternal(addon *assets.Addon, cmd command.Runner, data interface{}, enable bool) error {
	var err error

	if enable {
		for _, addon := range addon.Assets {
			var addonFile assets.CopyableFile
			if addon.IsTemplate() {
				addonFile, err = addon.Evaluate(data)
				if err != nil {
					return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
				}

			} else {
				addonFile = addon
			}
			if err := cmd.Copy(addonFile); err != nil {
				return errors.Wrapf(err, "enabling addon %s", addon.AssetName)
			}
		}
	} else {
		for _, addon := range addon.Assets {
			var addonFile assets.CopyableFile
			if addon.IsTemplate() {
				addonFile, err = addon.Evaluate(data)
				if err != nil {
					return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
				}

			} else {
				addonFile = addon
			}
			if err := cmd.Remove(addonFile); err != nil {
				return errors.Wrapf(err, "disabling addon %s", addon.AssetName)
			}
		}
	}
	return nil
}
