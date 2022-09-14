# Installation

The Flux Kluctl Controller requires an existing [Flux installation](https://fluxcd.io/docs/installation/) on the
same cluster that you plan to install the Flux Kluctl Controller to.

After Flux has been installed, you can install the Flux Kluctl Controller by running the following command:

```sh
kustomize build "https://github.com/kluctl/flux-kluctl-controller/config/install?ref=v0.6.1" | kubectl apply -f-
```


_NOTE: To set up Flux Alerts from KluctlDeployments you will need to patch the enum in the Alerts CRD.
There is a [patch](../config/patch/alerts-crd-patch.yaml) included in this repository that can do this for you. You can apply it directly or include the [yaml](../config/patch/alerts-crd-patch.yaml) version in `gotk-patch.yaml` with your `flux bootstrap`.
You can also add something like the following to your cluster's `kustomization.yaml`:_

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- gotk-components.yaml
- gotk-sync.yaml
patchesJson6902:
- target:
    group: apiextensions.k8s.io
    version: v1
    kind: CustomResourceDefinition
    name: alerts.notification.toolkit.fluxcd.io
  path: 'alerts-crd-patch.yaml' # The downloaded patch in your flux repository

```
