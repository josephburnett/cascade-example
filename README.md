# cascade-example
Example Kubernetes autoscaling setup

## Usage

Setup service chain:
```
make all
```

Run a load test:
```
make bench
```

Watch the show:
```
make dash
```

Stop the load test:
```
make nobench
```

Clean up:
```
make clean
```

## Prerequisites

1. Kubectl pointing to a GKE cluster.
2. [Stackdriver Metrics Adapter](https://github.com/GoogleCloudPlatform/k8s-stackdriver/blob/master/custom-metrics-stackdriver-adapter/README.md#configure-cluster) installed.
3. [Ko](https://github.com/google/ko) installed with `KO_DOCKER_REPO` pointing to the GCP project in which the GKE cluster resides (for container building).

## Disclaimers

This is the first thing that barely worked, not necessarily production best practices.  But it's a reasonable example of how to setup HPA with two metrics, CPU and a frontline service metric (qps) to reduce time-to-recovery.
