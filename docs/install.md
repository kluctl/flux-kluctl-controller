# Installation

The Flux Kluctl Controller can currently be either installed via Kustomize or via Helm.

## kustomize
You can install the Flux Kluctl Controller by running the following command:

```sh
kustomize build "https://github.com/kluctl/flux-kluctl-controller/config/install?ref=v0.9.0" | kubectl apply -f-
```

## Helm
A Helm Chart for the controller is also available [here](https://github.com/kluctl/charts/tree/main/charts/flux-kluctl-controller).
To install the controller via Helm, run:
```shell
$ helm repo add kluctl https://kluctl.github.io/charts
$ helm install flux-kluctl-controller kluctl/flux-kluctl-controller
```
