# msgbus

[![CI](https://github.com/leonkaihao/msgbus/actions/workflows/ci.yml/badge.svg)](https://github.com/leonkaihao/msgbus/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/leonkaihao/msgbus)](https://github.com/leonkaihao/msgbus)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A message bus implementation supporting NATS and MQTT protocols for Go applications.

## How to configure your dev env
# Prepare
1. Assume you have Docker installed.
2. Install VScode extensions: `Kubernetes`, `Kubernetes Kind`.
3. Install kubectl
   ```sh
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   ```
4. Install kind
   ```sh
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.12.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/
   ```
# Create K8s cluster
From VScode Activity bar, find Kubernetes, in the clouds view, right click Kind to create cluster.

# Build Go simulator images and push to K8s cluster
```sh
make build-all
make push-all
```
# Create deployments and services for NATs server, Consumer and Producer
```sh
make deploy-nats-stack
```
# Logging
```sh
kubectl log -f deployments/producer
kubectl log -f deployments/consumer
kubectl log -f deployments/nats
```
# Monitor NATs server:

## Monitor with nats-top
`kubectl port-forward deployments/nats 8222:8222`
Local:
`nats-top`
## Monitor with Grafana
### Deploy monitor stack
```sh
make deploy-nats-monitor
```
### Port-forward to host
```sh
kubectl port-forward deployments/grafana 3000:3000
```
### Config Grafana
1. Browser-open Grafana and login with admin:admin
2. `Configuration`-->`Data Sources`, add URL http://prometheus:9090 and save
3. `Create`-->`Import`-->input number `2279` for NATs dashboard-->`Load`.
Then you can enjoy NATs dashboard.
# Scaling
From `VScode Activity bar` --> `Kubernetes` --> `Clusters` --> `kind-kind` --> `Workloads` --> `Deployments` --> right click `consumer/nats/producer` --> Scale as you need.
If you need to remove one, just assign Scale=0

