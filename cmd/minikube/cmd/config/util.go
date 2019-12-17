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

package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/storageclass"
)

// defaultStorageClassProvisioner is the name of the default storage class provisioner
const defaultStorageClassProvisioner = "standard"

// Runs all the validation or callback functions and collects errors
func run(name string, value string, fns []setFn) error {
	var errors []error
	for _, fn := range fns {
		err := fn(name, value)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	}
	return nil
}

func findSetting(name string) (Setting, error) {
	for _, s := range settings {
		if name == s.name {
			return s, nil
		}
	}
	return Setting{}, fmt.Errorf("property name %q not found", name)
}

// Set Functions

// SetString sets a string value
func SetString(m config.MinikubeConfig, name string, val string) error {
	m[name] = val
	return nil
}

// SetMap sets a map value
func SetMap(m config.MinikubeConfig, name string, val map[string]interface{}) error {
	m[name] = val
	return nil
}

// SetConfigMap sets a config map value
func SetConfigMap(m config.MinikubeConfig, name string, val string) error {
	list := strings.Split(val, ",")
	v := make(map[string]interface{})
	for _, s := range list {
		v[s] = nil
	}
	m[name] = v
	return nil
}

// SetInt sets an int value
func SetInt(m config.MinikubeConfig, name string, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	m[name] = i
	return nil
}

// SetBool sets a bool value
func SetBool(m config.MinikubeConfig, name string, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	m[name] = b
	return nil
}

// EnableOrDisableStorageClasses enables or disables storage classes
func EnableOrDisableStorageClasses(name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrap(err, "Error parsing boolean")
	}

	class := defaultStorageClassProvisioner
	if name == "storage-provisioner-gluster" {
		class = "glusterfile"
	}
	storagev1, err := storageclass.GetStoragev1()
	if err != nil {
		return errors.Wrapf(err, "Error getting storagev1 interface %v ", err)
	}

	if enable {
		// Only StorageClass for 'name' should be marked as default
		err = storageclass.SetDefaultStorageClass(storagev1, class)
		if err != nil {
			return errors.Wrapf(err, "Error making %s the default storage class", class)
		}
	} else {
		// Unset the StorageClass as default
		err := storageclass.DisableDefaultStorageClass(storagev1, class)
		if err != nil {
			return errors.Wrapf(err, "Error disabling %s as the default storage class", class)
		}
	}

	return addons.EnableOrDisableAddon(name, val)
}

// ErrValidateProfile Error to validate profile
type ErrValidateProfile struct {
	Name string
	Msg  string
}

func (e ErrValidateProfile) Error() string {
	return e.Msg
}

// ValidateProfile checks if the profile user is trying to switch exists, else throws error
func ValidateProfile(profile string) (*ErrValidateProfile, bool) {

	validProfiles, invalidProfiles, err := config.ListProfiles()
	if err != nil {
		out.FailureT(err.Error())
	}

	// handling invalid profiles
	for _, invalidProf := range invalidProfiles {
		if profile == invalidProf.Name {
			return &ErrValidateProfile{Name: profile, Msg: fmt.Sprintf("%q is an invalid profile", profile)}, false
		}
	}

	profileFound := false
	// valid profiles if found, setting profileFound to trueexpectedMsg
	for _, prof := range validProfiles {
		if prof.Name == profile {
			profileFound = true
			break
		}
	}
	if !profileFound {
		return &ErrValidateProfile{Name: profile, Msg: fmt.Sprintf("profile %q not found", profile)}, false
	}
	return nil, true
}
