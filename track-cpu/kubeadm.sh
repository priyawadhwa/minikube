#!/bin/bash
set -ex

# minikube start --up-to-kubeadm --profile kubeadm

psrecord $(pgrep qemu-system) --plot plots/kubeadm.png --log logs/kubeadm.log --interval 1 --duration 60 --include-children

