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

func auxPath() string {
	return ""
}

// Auxilary downloads the auxiliary images tarball to the host
func Auxilary() error {
	// targetPath := auxPath()

	// if _, err := os.Stat(targetPath); err == nil {
	// 	glog.Infof("Found %s in cache, skipping download", targetPath)
	// 	return nil
	// }

	// // Make sure we support this k8s version
	// if !PreloadExists(k8sVersion, containerRuntime) {
	// 	glog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
	// 	return nil
	// }

	// out.T(style.FileDownload, "Downloading Kubernetes {{.version}} preload ...", out.V{"version": k8sVersion})
	// url := remoteTarballURL(k8sVersion, containerRuntime)

	// if err := download(url, targetPath); err != nil {
	// 	return errors.Wrapf(err, "download failed: %s", url)
	// }

	// if err := saveChecksumFile(k8sVersion, containerRuntime); err != nil {
	// 	return errors.Wrap(err, "saving checksum file")
	// }

	// if err := verifyChecksum(k8sVersion, containerRuntime, targetPath); err != nil {
	// 	return errors.Wrap(err, "verify")
	// }

	return nil
}
