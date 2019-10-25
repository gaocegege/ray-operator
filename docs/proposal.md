# Proposal: Running Ray Cluster on Kubernetes Natively with Autoscaler

Table of Contents
=================

   * [Proposal: Running Ray Cluster on Kubernetes Natively with Autoscaler](#proposal-running-ray-cluster-on-kubernetes-natively-with-autoscaler)
      * [Motivation](#motivation)
      * [Goal](#goal)
      * [Non-Goal](#non-goal)
      * [Design](#design)
         * [CRD](#crd)
         * [Workflow](#workflow)
            * [Create the Ray cluster](#create-the-ray-cluster)
            * [Update the Ray cluster](#update-the-ray-cluster)
         * [Limitations and Known Issues](#limitations-and-known-issues)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Motivation

Now ray has [autoscaler support for Kubernetes](https://ray.readthedocs.io/en/latest/autoscaling.html#kubernetes), which allows users deploy the ray cluster on Kubernetes. While it has some limitations using `ray up` to deploy:

- Users have to install ray and provide the kube-config to deploy.
- It is hard to create Ray clusters from client (kubectl or Kubernetes client-go)
- Users cannot know the status of the whole ray cluster on Kubernetes.
- It is hard to deploy multiple ray clusters with `ray up`.

## Goal

- Introduce Ray as a CRD, and allow users create/update/delete ray clusters.

## Non-Goal

- Implement autoscaler on Kubernetes, we reuse the Ray autoscaler.

## Design

### CRD

We provide two ways to configure a Ray cluster using one same CRD definition.

```go
// RaySpec defines the desired state of Ray
type RaySpec struct {
	MinWorkers         *int32                  `json:"minWorkers,omitempty"`
	MaxWorkers         *int32                  `json:"maxWorkers,omitempty"`
	InitialWorkers     *int32                  `json:"initialWorkers,omitempty"`
	AutoScalingMode    *ScalingMode            `json:"autoScalingMode,omitempty"`
	TargetUtilization  *int32                  `json:"targetUtilization,omitempty"`
	IdleTimeoutMinutes *int32                  `json:"idleTimeoutMinutes,omitempty"`
	Head               *corev1.PodTemplateSpec `json:"head,omitempty"`
	Worker             *corev1.PodTemplateSpec `json:"worker,omitempty"`
	// TODO(gaocegege): Add other configurations.

	Configuration *string `json:"configuration,omitempty"`
}

// RayStatus defines the observed state of Ray
type RayStatus struct {
	Launcher ReplicaStatus `json:"launcher,omitempty"`
	Head     ReplicaStatus `json:"head,omitempty"`
	Worker   ReplicaStatus `json:"worker,omitempty"`
	// Conditions is an array of current observed ray conditions.
	Conditions []RayCondition `json:"conditions,omitempty"`

	// Represents time when the ray was acknowledged by the ray operator.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The generation observed by the ray operator.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents last time when the Ray was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

// ReplicaStatus is the status field for the replica.
type ReplicaStatus struct {
	// Total number of non-terminated pods targeted by this replica (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of non-terminated pods targeted by this replica that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// Total number of ready pods targeted by this replica.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this replica.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Total number of unavailable pods targeted by this replica. This is the total number of
	// pods that are still required for the replica to have 100% available capacity. They may
	// either be pods that are running but not yet available or pods that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`
}
```

One option is to define ray autoscaler configuration with CRD (which is recommended):

```yaml
apiVersion: ray.k8s.io/v1alpha1
kind: Ray
metadata:
  name: ray-sample
  namespace: ray
spec:
  configuration: |-
    # An unique identifier for the head node and workers of this cluster.
    cluster_name: default

    # The minimum number of workers nodes to launch in addition to the head
    # node. This number should be >= 0.
    min_workers: 0

    # The maximum number of workers nodes to launch in addition to the head
    # node. This takes precedence over min_workers.
    max_workers: 2

    # Kubernetes resources that need to be configured for the autoscaler to be
    # able to manage the Ray cluster. If any of the provided resources don't
    # exist, the autoscaler will attempt to create them. If this fails, you may
    # not have the required permissions and will have to request them to be
    # created by your cluster administrator.
    provider:
        type: kubernetes
    ...
```

Other approach is to define Kubernetes style configuration:

```yaml
apiVersion: ray.k8s.io/v1alpha1
kind: Ray
metadata:
  name: ray-sample
  namespace: ray
spec:
  minWorkers: 1
  maxWorkers: 2
  initialWorkers: 1
  autoScalingMode: default
  targetUtilization: 80
  idleTimeoutMinutes: 60
```

It is more friendly for users which use kubernetes client to create Ray clusters.

### Workflow

#### Create the Ray cluster

When users create the Ray CRD, the controller will create a launcher to launch the cluster. The launcher is similar to [`horovodrun`](https://github.com/horovod/horovod/blob/master/docs/running.rst) or `spark-submit` but for the Ray cluster:

```yaml
metadata:
  creationTimestamp: "2019-10-25T02:18:13Z"
  generateName: ray-sample-546685944-
  labels:
    ray: ray-sample
    ray-launcher: ray-sample
  name: ray-sample-546685944-2sx9k
  namespace: ray
spec:
  containers:
  - command:
    - bash
    - -c
    - echo yes | ray up /mounted-configmap/example-full.yaml
    image: rayproject/autoscaler
    imagePullPolicy: IfNotPresent
    name: ray-launcher
```

The configuration of Ray autoscaler will be mounted into the launcher, and used to create Ray clusters on the same Kubernetes cluster. The launcher will be finished when the cluster is setup.

```
$ kubectl --namespace ray get pods
ray-head-zlqwh                   1/1     Running       0          3h17m
ray-worker-asfla                 1/1     Running       0          3h17m
ray-launcher-afhys               1/1     Completed     0          3h17m
$ kubectl --namespace ray get Ray -o json | jq ".items[].status"
{
    launcher: {
        conditions: [
            {
                {
                    "lastProbeTime": null,
                    "lastTransitionTime": "2019-10-25T02:18:13Z",
                    "status": "True",
                    "type": "SuccessfulLaunch"
                },
            }
        ]
    }
    head: {
        replicas: 1,
        availableReplicas: 1,
    }
    worker: {
        replicas: 1,
        availableReplicas: 1,
    }
}
```

#### Update the Ray cluster

We can use kubectl to update it.

### Limitations and Known Issues

- Not sure if `ray up` has any daemon processes. AFAIK, I do not find such a process.
