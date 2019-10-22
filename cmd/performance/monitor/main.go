package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/performance/monitor"
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	client := monitor.NewClient(context.Background(), "kubernetes", "minikube")
	prs, err := client.ListOpenPRsWithLabel("")
	if err != nil {
		return errors.Wrap(err, "listing open prs")
	}
	// TODO: priyawadhwa@ for each PR we should comment the error if we get one?
	for _, pr := range prs {
		newCommitsExist, err := client.NewCommitsExist(pr, "minikube-bot")
		if err != nil {
			return err
		}
		if !newCommitsExist {
			continue
		}
		// TODO: priyawadhwa@ we should download mkcmp for each run?

	}
	return nil
}
