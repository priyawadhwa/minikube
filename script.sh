#!/bin/bash


while true; do


minikube stop
minikube start
STR=$(minikube ssh sudo systemctl status kubelet)
SUB='node-ip=192.168.64.3'
if [[ "$STR" == *"$SUB"* ]]; then
  echo "It's there."
else
   echo "Node ip is different"
   exit 1

fi

done

