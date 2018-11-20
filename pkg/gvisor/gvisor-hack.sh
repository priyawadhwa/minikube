#!/usr/bin/env bash
pushd $(mktemp -d)
until ping -c 1 storage.googleapis.com &> /dev/null; do echo "waiting for dns..."; sleep 1; done
wget http://storage.googleapis.com/balintp-minikube/gvisor-containerd-shim
sudo mv gvisor-containerd-shim /usr/bin/gvisor-containerd-shim

sudo chmod +x /usr/bin/gvisor-containerd-shim
sudo mkdir -p /usr/local/bin
mkdir -p /tmp/runsc
sudo mkdir -p /run/containerd/runsc

wget http://storage.googleapis.com/gvisor/releases/nightly/latest/runsc
chmod a+x runsc
sudo mv runsc /usr/local/bin
popd
