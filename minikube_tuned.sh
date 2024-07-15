#!/bin/sh

minikube start --extra-config=kubelet.sync-frequency=10s
