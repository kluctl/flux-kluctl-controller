<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: v1alpha1 specs
linkTitle: v1alpha1 specs
description: flux.kluctl.io/v1alpha1 documentation
weight: 10
---
-->

# flux.kluctl.io/v1alpha1

This is the v1alpha1 API specification for defining continuous delivery pipelines
of Kluctl Deployments.

## Specification

- [KluctlDeployment CRD](kluctldeployment.md)
    + [Spec fields](kluctldeployment.md#spec-fields)
    + [Reconciliation](kluctldeployment.md#reconciliation)
    + [Kubeconfigs and RBAC](kluctldeployment.md#kubeconfigs-and-rbac)
    + [Git authentication](kluctldeployment.md#git-authentication)
    + [Helm Repository authentication](kluctldeployment.md#helm-repository-authentication)
    + [Secrets Decryption](kluctldeployment.md#secrets-decryption)
    + [Status](kluctldeployment.md#status)

## Implementation

* [flux-kluctl-controller](https://github.com/kluctl/flux-kluctl-controller)
