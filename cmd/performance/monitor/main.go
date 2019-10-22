package main

import (
	"context"
	"fmt"
	"os"

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
	fmt.Println(client.TimeOfLastComment(5694, "minikube-bot"))
	return nil
}
