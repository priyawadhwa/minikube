#!/bin/bash
set -ex

# minikube start --up-to-vm-creation --profile vm



psrecord $(pgrep qemu-system) --plot plots/vm.png --log logs/vm.log --interval 1 --duration 60 --include-children

