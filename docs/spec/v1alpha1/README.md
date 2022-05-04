# kustomize.toolkit.fluxcd.io/v1alpha1

This is the v1alpha1 API specification for defining continuous delivery pipelines
of Kluctl Deployments.

## Specification

- [KluctlDeployment CRD](kluctldeployment.md)
    + [Source reference](kluctldeployment.md#source-reference)
    + [Reconciliation](kluctldeployment.md#reconciliation)
    + [Pruning](kluctldeployment.md#pruning)
    + [Kubeconfigs and RBAC](kluctldeployment.md#kubeconfigs-and-rbac)
    + [Status](kluctldeployment.md#status)

## Implementation

* [flux-kluctl-controller](https://github.com/kluctl/flux-kluctl-controller)
