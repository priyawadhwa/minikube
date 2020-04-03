/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/perf"
)

var (
	logsFile string
)

func init() {
	rootCmd.Flags().StringVar(&logsFile, "logs-file", "", "Path to a file with logs that need to be tracked.")
	flag.Parse()
}

var rootCmd = &cobra.Command{
	Use:           "mkcmp [path to first binary] [path to second binary]",
	Short:         "mkcmp is used to compare performance of two minikube binaries",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validateArgs(args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return perf.CompareMinikubeStart(context.Background(), os.Stdout, args, logsFile)
	},
}

func validateArgs(args []string) error {
	if len(args) != 2 {
		return errors.New("mkcmp requires two minikube binaries to compare: mkcmp [path to first binary] [path to second binary]")
	}
	if logsFile == "" {
		return errors.New("Please pass in a path to a file containing logs to time via --logs-file")
	}
	if _, err := os.Stat(logsFile); err != nil {
		return fmt.Errorf("Please pass in a valid path via --logs-file: %v", err)
	}
	return nil
}

// Execute runs the mkcmp command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
