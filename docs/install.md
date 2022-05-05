# Installation

The Flux Kluctl Controller requires an existing [Flux installation](https://fluxcd.io/docs/installation/) on the
same cluster that you plan to install the Flux Kluctl Controller to.

After Flux has been installed, you can install the Flux Kluctl Controller by running the following command:

```sh
kustomize build "https://github.com/kluctl/flux-kluctl-controller//config/install?ref=v0.1.4" | kubectl apply -f-
```
