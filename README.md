<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: "Flux Kluctl Controller"
linkTitle: "Flux Kluctl Controller"
description: "Flux Kluctl Controller documentation."
weight: 200
---
-->

# Flux Kluctl Controller

The Flux Kluctl Controller is a Kubernetes operator, specialized in running 
continuous delivery pipelines for infrastructure defined with [kluctl](https://kluctl.io).

## Motivation

[kluctl](https://kluctl.io) is a tool that allows you to declare and manage small, large, simple
and/or complex multi-env and multi-cluster deployments. It is designed in a way that allows seamless
co-existence of CLI centered DevOps and automation, for example in the form of GitOps/flux.

This means that you can continue doing local development of your deployments and test them from your local machine,
for example by regularly running [kluctl diff](https://kluctl.io/docs/kluctl/reference/commands/diff/). When you believe
you're done with your work, you can then commit your changes to Git and let the Flux Kluctl Controller do the
actual deployment.

You could also have a dedicated [target](https://kluctl.io/docs/kluctl/reference/kluctl-project/targets/)
that you solely use for local development and deployment testing and then let the Flux Kluctl Controller handle
the deployments to the real (e.g. pre-prod or prod) targets.

This way you can have both:
1. Easy and reliable development and testing of your deployments (no more change+commit+push+wait+error+retry cycles).
2. Consistent GitOps style automation.

The Flux Kluctl Controller supports all types of Kluctl projects, including simple ones where a single Git repository
contains all necessary data and complex ones where for example clusters or target configurations are in other Git
repositories.

## Installation

Installation instructions can be found [here](./docs/install.md)

## Design

The reconciliation process can be defined with a Kubernetes custom resource
that describes a pipeline such as:
- **fetch** root kluctl project from source-controller (Git repository or S3 bucket)
- **deploy** the specified target via [kluctl deploy](https://kluctl.io/docs/kluctl/reference/commands/deploy/)
- **prune** orphaned objects via [kluctl prune](https://kluctl.io/docs/kluctl/reference/commands/prune/)
- **validate** the deployment status via [kluctl validate](https://kluctl.io/docs/kluctl/reference/commands/validate/)
- **alert** if something went wrong
- **notify** if the cluster state changed

A pipeline runs on-a-schedule and ca be triggered manually by a
cluster admin or automatically by a source event such as a Git revision change.

When a pipeline is removed from the cluster, the controller's GC terminates
all the objects previously created by that pipeline.

A pipeline can be suspended, while in suspension the controller
stops the scheduler and ignores any source events.
Deleting a suspended pipeline does not trigger garbage collection.

Alerting can be configured with a Kubernetes custom resource
that specifies a webhook address, and a group of pipelines to be monitored.

The API design of the controller can be found at [kluctldeployment.flux.kluctl.io/v1beta1](v1alpha1/README.md).

## Example

After installing flux-kluctl-controller, we can create a Kluctl deployment that automatically deploys the
[Microservices Demo](https://kluctl.io/docs/guides/tutorials/microservices-demo/3-templating-and-multi-env/).

Create a KluctlDeployment that uses the demo project source to deploy the `test` target to the same cluster that the
controller runs on.

```yaml
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: microservices-demo-test
  namespace: flux-system
spec:
  interval: 10m
  source:
    url: https://github.com/kluctl/kluctl-examples.git
    path: "./microservices-demo/3-templating-and-multi-env/"
  timeout: 2m
  target: test
  context: default
  prune: true
```

This example will deploy a fully-fledged microservices application with multiple backend services, frontends and
databases, all via one single `KluctlDeployment`.

To deploy the same Kluctl project to another target (e.g. prod), simply create the following resource.

```yaml
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: microservices-demo-prod
  namespace: flux-system
spec:
  interval: 10m
  source:
    url: https://github.com/kluctl/kluctl-examples.git
    path: "./microservices-demo/3-templating-and-multi-env/"
  timeout: 2m
  target: prod
  context: default
  prune: true
```
