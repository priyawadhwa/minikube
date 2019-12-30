package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/performance/monitor"
)

func main() {

	for {
		log.Print("~~~~~~~~~ Starting performance analysis ~~~~~~~~~~~~~~")
		if err := execute(context.Background()); err != nil {
			log.Printf("Error executing: %v", err)
		}
		time.Sleep(10 * time.Minute)
	}

}

func execute(ctx context.Context) error {
	logsFile := "/home/performance-monitor/logs.txt"
	if _, err := os.Stat(logsFile); err != nil {
		return err
	}
	client := monitor.NewClient(context.Background(), "kubernetes", "minikube")
	prs, err := client.ListOpenPRsWithLabel("")
	if err != nil {
		return errors.Wrap(err, "listing open prs")
	}
	log.Print("got prs:", prs)
	// TODO: priyawadhwa@ for each PR we should comment the error if we get one?
	for _, pr := range prs {
		log.Printf("~~~ Analyzing PR %d ~~~", pr)
		newCommitsExist, err := client.NewCommitsExist(pr, "minikube-bot")
		if err != nil {
			return err
		}
		if !newCommitsExist {
			log.Println("New commits don't exist, skipping rerun...")
			continue
		}
		// TODO: priyawadhwa@ we should download mkcmp for each run?
		var message string
		message, err = monitor.RunMkcmp(ctx, pr)
		if err != nil {
			message = fmt.Sprintf("Error: %v\n%s", err, message)
		}
		log.Printf("got message for pr %d:\n%s\n", pr, message)
		if err := client.CommentOnPR(pr, message); err != nil {
			return err
		}
		log.Print("successfully commented on PR")
	}
	return nil
}
