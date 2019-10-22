/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/performance"
)

var (
	quiet bool
)

var rootCmd = &cobra.Command{
	Use:   "mkcmp [path to first binary] [path to second binary]",
	Short: "mkcmp is used to compare performance of two minikube binaries",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validateArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		binaries, err := getBinaries(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var out io.Writer
		out = os.Stdout
		if quiet {
			out = ioutil.Discard
		}
		if err := performance.CompareMinikubeStart(context.Background(), out, binaries); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func validateArgs(args []string) error {
	if len(args) != 2 {
		return errors.New("mkcmp requries two minikube binaries to compare: mkcmp [path to first binary] [path to second binary]")
	}
	return nil
}

func getBinaries(args []string) ([]*performance.Binary, error) {
	var binaries []*performance.Binary
	for _, a := range args {
		b, err := performance.NewBinary(a)
		if err != nil {
			return nil, err
		}
		binaries = append(binaries, b)
	}
	return binaries, nil
}

// Execute runs the mkcmp command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "", false, "only output results")
}
