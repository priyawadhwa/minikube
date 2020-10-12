/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package download

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	// PreloadVersion is the current version of the preloaded tarball
	PreloadVersion = "v6"
	// AuxVersion is all of the auxilary images required for minikube
	// NOTE: You need to bump this version up when upgrading auxiliary docker images
	AuxVersion = "v1"
	// PreloadBucket is the name of the GCS bucket where preloaded volume tarballs exist
	PreloadBucket = "minikube-preloaded-volume-tarballs"
)

// PreloadName returns name of the preload tarball
func PreloadName(k8sVersion, containerRuntime string) string {
	if containerRuntime == "crio" {
		containerRuntime = "cri-o"
	}
	var storageDriver string
	if containerRuntime == "cri-o" {
		storageDriver = "overlay"
	} else {
		storageDriver = "overlay2"
	}
	return fmt.Sprintf("preloaded-images-k8s-%s-%s-%s-%s-%s.tar.lz4", PreloadVersion, k8sVersion, containerRuntime, storageDriver, runtime.GOARCH)
}

// returns the name of the checksum file
func checksumName(tarballName string) string {
	return fmt.Sprintf("%s.checksum", tarballName)
}

// returns target dir for all cached items related to preloading
func targetDir() string {
	return localpath.MakeMiniPath("cache", "preloaded-tarball")
}

// ChecksumPath returns the local path to the cached checksum file
func ChecksumPath(tarballName string) string {
	return filepath.Join(targetDir(), tarballName)
}

// TarballPath returns the local path to the cached tarball
func TarballPath(tarballName string) string {
	return filepath.Join(targetDir(), tarballName)
}

// remoteTarballURL returns the URL for the remote tarball in GCS
func remoteTarballURL(tarballName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", PreloadBucket, tarballName)
}

// TarballExists returns true if there is a tarball at the specified path that can be used
func TarballExists(tarballName string, forcePreload ...bool) bool {
	// TODO (#8166): Get rid of the need for this and viper at all
	force := false
	if len(forcePreload) > 0 {
		force = forcePreload[0]
	}

	// TODO: debug why this func is being called two times
	glog.Info("Checking if preload exists...")
	if !viper.GetBool("preload") && !force {
		return false
	}

	targetPath := TarballPath(tarballName)
	// Omit remote check if tarball exists locally
	if _, err := os.Stat(targetPath); err == nil {
		glog.Infof("Found local preload: %s", targetPath)
		return true
	}

	url := remoteTarballURL(tarballName)
	resp, err := http.Head(url)
	if err != nil {
		glog.Warningf("%s fetch error: %v", url, err)
		return false
	}

	// note: err won't be set if it's a 404
	if resp.StatusCode != 200 {
		glog.Warningf("%s status code: %d", url, resp.StatusCode)
		return false
	}

	glog.Infof("Found remote preload: %s", url)
	return true
}

// Preload caches the preloaded images tarball on the host machine
func Preload(k8sVersion, containerRuntime string) error {
	essentialName := PreloadName(k8sVersion, containerRuntime)
	auxName := auxName(containerRuntime)
	tarballNames := []string{essentialName, auxName}

	var needToDownload []string
	for _, name := range tarballNames {
		tp := TarballPath(name)
		if _, err := os.Stat(tp); err != nil {
			glog.Infof("Didn't find %s in cache, will have to download\n", tp)
			needToDownload = append(needToDownload, name)
		}
	}
	if needToDownload == nil {
		glog.Infof("All preload tarballs are cached, skipping download")
		return nil
	}

	out.T(style.FileDownload, "Downloading Kubernetes {{.version}} preload ...", out.V{"version": k8sVersion})
	for _, tarballName := range needToDownload {
		tarballName := tarballName
		tp := TarballPath(tarballName)
		if _, err := os.Stat(tp); err == nil {
			glog.Infof("Found %s in cache, skipping download", tp)
			continue
		}
		// Make sure we support this k8s version
		if !TarballExists(tarballName) {
			glog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
			continue
		}

		url := remoteTarballURL(tarballName)

		if err := download(url, tp); err != nil {
			return errors.Wrapf(err, "download failed: %s", url)
		}

		if err := saveChecksumFile(tarballName); err != nil {
			return errors.Wrap(err, "saving checksum file")
		}

		if err := verifyChecksum(tarballName, tp); err != nil {
			return errors.Wrap(err, "verify")
		}
	}

	return nil
}

func saveChecksumFile(tarballName string) error {
	glog.Infof("saving checksum for %s ...", tarballName)
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return errors.Wrap(err, "getting storage client")
	}
	attrs, err := client.Bucket(PreloadBucket).Object(tarballName).Attrs(ctx)
	if err != nil {
		return errors.Wrap(err, "getting storage object")
	}
	checksum := attrs.MD5
	return ioutil.WriteFile(ChecksumPath(tarballName), checksum, 0o644)
}

// verifyChecksum returns true if the checksum of the local binary matches
// the checksum of the remote binary
func verifyChecksum(tarballName, path string) error {
	glog.Infof("verifying checksumm of %s ...", path)
	// get md5 checksum of tarball path
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "reading tarball")
	}
	checksum := md5.Sum(contents)

	remoteChecksum, err := ioutil.ReadFile(ChecksumPath(tarballName))
	if err != nil {
		return errors.Wrap(err, "reading checksum file")
	}

	// create a slice of checksum, which is [16]byte
	if string(remoteChecksum) != string(checksum[:]) {
		return fmt.Errorf("checksum of %s does not match remote checksum (%s != %s)", path, string(remoteChecksum), string(checksum[:]))
	}
	return nil
}
