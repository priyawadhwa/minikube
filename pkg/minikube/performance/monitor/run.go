package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"bytes"
	"github.com/pkg/errors"
)
// RunMkcmp runs minikube built at the given pr against minikube at master
// TODO: priyawadhwa@ to figure this out.
func RunMkcmp(ctx context.Context, pr int) (string, error) {
	cmd := exec.CommandContext(ctx, "mkcmp", "minikube", fmt.Sprintf("pr://%d", pr))
	stdOut := bytes.NewBuffer([]byte{})
	stdErr := bytes.NewBuffer([]byte{})
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		return stdErr.String(), errors.Wrap(err, "running mkcmp")
	}
	return stdOut.String(), nil
}  