---
title: "completion"
linkTitle: "completion"
weight: 1
date: 2019-08-01
description: >
  Outputs minikube shell completion for the given shell (bash or zsh)
---


### Overview

Outputs minikube shell completion for the given shell (bash or zsh)

This depends on the bash-completion binary.  Example installation instructions:

### Usage

```
minikube completion SHELL [flags]
```

## Example: macOS

```shell
brew install bash-completion
source $(brew --prefix)/etc/bash_completion
minikube completion bash > ~/.minikube-completion  # for bash users
$ minikube completion zsh > ~/.minikube-completion  # for zsh users
$ source ~/.minikube-completion
```

## Example: Ubuntu

```shell
apt-get install bash-completion
source /etc/bash-completion
source <(minikube completion bash) # for bash users
source <(minikube completion zsh) # for zsh users
```

Additionally, you may want to output the completion to a file and source in your .bashrc

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```
