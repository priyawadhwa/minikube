package monitor

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

// RunMkcmp runs minikube built at the given pr against minikube at master
// TODO: priyawadhwa@ to figure this out.
func RunMkcmp(ctx context.Context, pr int) (string, error) {
	if err := buildMinikubeAtHead(ctx); err != nil {
		return "", errors.Wrap(err, "building minikube at head")
	}
	minikubeAtHead := filepath.Join(minikubeDir(), "out/minikube")

	cmd := exec.CommandContext(ctx, "mkcmp", minikubeAtHead, fmt.Sprintf("pr://%d", pr), "--quiet")
	stdOut := bytes.NewBuffer([]byte{})
	stdErr := os.Stderr
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	log.Print("running mkcmp: ", cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "running mkcmp")
	}
	return stdOut.String(), nil
}

func buildMinikubeAtHead(ctx context.Context) error {
	log.Print("building minikube at head in", minikubeDir())
	gitPull := exec.CommandContext(ctx, "git", "pull", "origin", "master")
	gitPull.Dir = minikubeDir()

	if err := runCmd(gitPull); err != nil {
		return errors.Wrap(err, "updating minikube master branch")
	}

	makeMinikube := exec.CommandContext(ctx, "make")
	makeMinikube.Dir = minikubeDir()

	if err := runCmd(makeMinikube); err != nil {
		return errors.Wrap(err, "building minikube via make")
	}
	return nil
}

func minikubeDir() string {
	return filepath.Join(os.Getenv("HOME"), "minikube")
}

func runCmd(cmd *exec.Cmd) error {
	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	cmd.Stderr = buf

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running %v in %s:\n%s", cmd.Args, cmd.Dir, buf.String())
	}
	return nil
}
