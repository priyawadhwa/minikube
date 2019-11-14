#!/bin/bash
set -ex

psrecord $(pgrep qemu-system) --plot plots/minikube.png --log logs/minikube.log --interval 1 --duration 60 --include-children

