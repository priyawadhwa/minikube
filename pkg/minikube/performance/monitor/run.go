package monitor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"bytes"
	"github.com/pkg/errors"
)
// RunMkcmp runs minikube built at the given pr against minikube at master
// TODO: priyawadhwa@ to figure this out.
func RunMkcmp(ctx context.Context, pr int) (string, error) {
	if err := buildMinikubeAtHead(); err != nil {
		return "", errors.Wrap(err, "building minikube at head")
	}
	minikubeAtHead := filepath.Join(minikubeDir(), "out/minikube")
	
	cmd := exec.CommandContext(ctx, "mkcmp", minikubeAtHead, fmt.Sprintf("pr://%d", pr), "--quiet")
	stdOut := bytes.NewBuffer([]byte{})
	stdErr := os.Stderr
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
 
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "running mkcmp")
	}
	return stdOut.String(), nil
}  

func buildMinikubeAtHead(ctx context.Context) error {
	gitPull := exec.CommandContext(ctx, "git", "pull", "origin", "master")
	gitPull.Dir = minikubeDir()

	if err := gitPull.Run(); err != nil {
		return errors.Wrap(err, "updating minikube master branch")
	}

	makeMinikube := exec.CommandContext(ctx, "make")
	makeminikube.Dir = minikubeDir()

	if err := makeMinikube.Run(); err != nil {
		return errors.Wrap(err, "building minikube via make")
	}
	return nil
}

func minikubeDir() string {
	return filepath.Join(os.Getenv("HOME"), "minikube")
}