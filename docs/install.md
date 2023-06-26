<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: Installation
description: Installation documentation
weight: 10
---
-->

# Installation

> ⚠️**The flux-kluctl-controller is deprecated**⚠️
>
> Please migrate to the new [Kluctl Controller](https://kluctl.io/docs/kluctl/reference/gitops/)
>
> The migration guide can be found [here](https://kluctl.io/docs/kluctl/reference/gitops/migration/)


The Flux Kluctl Controller can currently be either installed via Kustomize or via Helm.

## kustomize
You can install the Flux Kluctl Controller by running the following command:

```sh
kustomize build "https://github.com/kluctl/flux-kluctl-controller/config/install?ref=v0.16.4" | kubectl apply -f-
```

## Helm

> ⚠️**The flux-kluctl-controller is deprecated**⚠️
>
> New Helm Charts will not be released!


A Helm Chart for the controller is also available [here](https://github.com/kluctl/charts/tree/main/charts/flux-kluctl-controller).
To install the controller via Helm, run:
```shell
$ helm repo add kluctl https://kluctl.github.io/charts
$ helm install flux-kluctl-controller kluctl/flux-kluctl-controller
```
