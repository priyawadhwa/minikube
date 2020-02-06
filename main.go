package main

import (
	"fmt"
	"os"
	"os/exec"

	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
)

const (
	profile = "generate-preloaded-images-tar"
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	// TODO: pass in flag for version here
	images, err := images.Kubeadm("", "v1.17.0")
	if err != nil {
		return err
	}
	fmt.Println(images)

	return nil
}

func downloadImages(version string) error {
	args := []string{
		fmt.Sprintf("-p=%s", profile),
		fmt.Sprintf("--kubernetes-version=%s", version),
		fmt.Sprintf("--memory=6000", version),
	}
	cmd.RunStart(nil, args)
	defer deleteMinikube(profile)

	if err := createTarOfImages(profile); err != nil {
		return err
	}

	if err := copyTarOfImages(profile); err != nil {
		return err
	}

	return nil
}

func startMinikube() error {

}

func deleteMinikube() error {
	cmd := exec.Command("minikube", "delete", "-p", profile)
	return cmd.Run()
}

// SSH into the minikube VM and create a tarball of the docker directory
func createTarOfImages() error {
	return nil
}

// Copy tarball of docker images onto host machine
func copyTarOfImages() error {
	return nil
}
