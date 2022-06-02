# KluctlDeployment

The `KluctlDeployment` API defines a deployment of a [target](https://kluctl.io/docs/reference/kluctl-project/targets/)
from a [Kluctl Project](https://kluctl.io/docs/reference/kluctl-project/).

## Example

```yaml
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: microservices-demo
spec:
  interval: 1m
  url: https://github.com/kluctl/kluctl-examples.git
  ref:
    branch: main
---
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: microservices-demo-prod
spec:
  interval: 5m
  path: "./microservices-demo/3-templating-and-multi-env/"
  sourceRef:
    kind: GitRepository
    name: microservices-demo
  timeout: 2m
  target: prod
  prune: true
  # kluctl targets specify the expected context name, which does not necessarily match the context name
  # found while it is deployed via the controller. This means we must pass a kubeconfig to kluctl that has the
  # context renamed to the one that it expects.
  renameContexts:
    - oldContext: default
      newContext: kind-kind
```

In the above example, two objects are being created, a GitRepository that points to the Kluctl project and KluctlDeployment
that defines the deployment based on the Kluctl project.

The deployment is performed every 5 minutes or whenever the source changes. It will deploy the `prod`
[target](https://kluctl.io/docs/reference/kluctl-project/targets/) and then prune orphaned objects afterwards.

It uses the `default` context provided by the default Flux service account and rename it to `kind-kind` so that it is
compatible with the context specified in the example's `prod` target.

## Source reference

The KluctlDeployment `spec.sourceRef` is a reference to an object managed by
[source-controller](https://github.com/fluxcd/source-controller). When the source
[revision](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/common.md#source-status)
changes, it generates a Kubernetes event that triggers a reconciliation attempt.

Source supported types:

* [GitRepository](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/gitrepositories.md)
* [Bucket](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/buckets.md)

The Kluctl project found in the referenced source might also internally reference other Git repositories, for example
by loading variables from Git repositories or including other Git repositories in your deployments. In this case,
the controller will re-use the credentials from the root project's GitRepository for further authentication.

`spec.path` specifies the subdirectory inside the referenced source to be used as the project root.

## Target
`spec.target` specifies the target to be deployed. It must exist in the Kluctl projects
[kluctl.yaml targets](https://kluctl.io/docs/reference/kluctl-project/targets/) list.

## Reconciliation

The KluctlDeployment `spec.interval` tells the controller at which interval to try reconciliations.
The interval time units are `s`, `m` and `h` e.g. `interval: 5m`, the minimum value should be over 60 seconds.

The KluctlDeployment reconciliation can be suspended by setting `spec.suspend` to `true`.

The controller can be told to reconcile the KluctlDeployment outside of the specified interval
by annotating the KluctlDeployment object with `fluxcd.io/reconcileAt`.

On-demand execution example:

```bash
kubectl annotate --overwrite kluctldeployment/microservices-demo-prod fluxcd.io/reconcileAt="$(date +%s)"
```

## Pruning

To enable pruning, set `spec.prune` to `true`. This will cause the controller to run `kluctl prune` after each
successful deployment.

## Kluctl Options
The [kluctl deploy](https://kluctl.io/docs/reference/commands/deploy/) command has multiple arguments that influence
how the deployment is performed. `KluctlDeployment`'s can set most of these arguements as well:

### args
`spec.args` is a map of strings representing [arguments](https://kluctl.io/docs/reference/deployments/deployment-yml/#args)
passed to the deployment. Example:

```yaml
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: example
spec:
  interval: 5m
  sourceRef:
    kind: GitRepository
    name: example
  timeout: 2m
  target: prod
  args:
    arg1: value1
    arg2: value2
```

The above example is equivalent to calling `kluctl deploy -t prod -a arg1=value1 -a arg2=value2`.

### updateImages
`spec.updateImages` is a boolean that specifies whether images used via
[`image.get_image(...)`](https://kluctl.io/docs/reference/deployments/images/#imagesget_image) should use the latest
image found in the registry.

This is equivalent to calling `kluctl deploy -t prod -u`

### images
`spec.images` specifies a list of fixed images to be used by
[`image.get_image(...)`](https://kluctl.io/docs/reference/deployments/images/#imagesget_image). Example:

```
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: example
spec:
  interval: 5m
  sourceRef:
    kind: GitRepository
    name: example
  timeout: 2m
  target: prod
  images:
    - image: nginx
      resultImage: nginx:1.21.6
      namespace: example-namespace
      deployment: Deployment/example
    - image: registry.gitlab.com/my-org/my-repo/image
      resultImage: registry.gitlab.com/my-org/my-repo/image:1.2.3
```

The above example will cause the `images.get_image("nginx")` invocations of the `example` Deployment to return
`nginx:1.21.6`. It will also cause all `images.get_image("registry.gitlab.com/my-org/my-repo/image")` invocations
to return `registry.gitlab.com/my-org/my-repo/image:1.2.3`.

The fixed images provided here take precedence over the ones provided in the
[target definition](https://kluctl.io/docs/reference/kluctl-project/targets/#images).

`spec.images` is equivalent to calling `kluctl deploy -t prod --fixed-image=nginx:example-namespace:Deployment/example=nginx:1.21.6 ...`
and to `kluctl deploy -t prod --fixed-images-file=fixed-images.yaml` with `fixed-images.yaml` containing:

```yaml
images:
- image: nginx
  resultImage: nginx:1.21.6
  namespace: example-namespace
  deployment: Deployment/example
- image: registry.gitlab.com/my-org/my-repo/image
  resultImage: registry.gitlab.com/my-org/my-repo/image:1.2.3
```

It is advised to use [dynamic targets](https://kluctl.io/docs/reference/kluctl-project/targets/dynamic-targets/)
instead of providing images directly in the ´KluctlDeployment` object.

### dryRun
`spec.dryRun` is a boolean value that turns the deployment into a dry-run deployment. This is equivalent to calling
`kluctl deploy -t prod --dry-run`.


### noWait
`spec.noWait` is a boolean value that disables all internal waiting (hooks and readiness). This is equivalent to calling
`kluctl deploy -t prod --no-wait`.

### forceApply
`spec.forceApply` is a boolean value that causes kluctl to solve conflicts via force apply. This is equivalent to calling
`kluctl deploy -t prod --force-apply`.

### replaceOnError and forceReplaceOnError
`spec.replaceOnError` and `spec.forceReplaceOnError` are both boolean values that cause kluctl to perform a replace
after a failed apply. `forceReplaceOnError` goes a step further and deletes and recreates the object in question.
These are equivalent to calling `kluctl deploy -t prod --replace-on-error` and `kluctl deploy -t prod --force-replace-on-error`.

### abortOnError
`spec.abortOnError` is a boolean value that causes kluctl to abort as fast as possible in case of errors. This is equivalent to calling
`kluctl deploy -t prod --abort-on-error`.

### includeTags, excludeTags, includeDeploymentDirs and excludeDeploymentDirs
`spec.includeTags` and `spec.excludeTags` are lists of tags to be used in inclusion/exclusion logic while deploying.
These are equivalent to calling `kluctl deploy -t prod --include-tag <tag1>` and `kluctl deploy -t prod --exclude-tag <tag2>`.

`spec.includeDeploymentDirs` and `spec.excludeDeploymentDirs` are lists of relative deployment directories to be used in
inclusion/exclusion logic while deploying. These are equivalent to calling `kluctl deploy -t prod --include-tag <tag1>`
and `kluctl deploy -t prod --exclude-tag <tag2>`.

## Dependencies

KluctlDeployment's can specify dependencies via `spec.dependsOn`. The functionality is equivalent to
[kustomization-dependencies](https://fluxcd.io/docs/components/kustomize/kustomization/#kustomization-dependencies)
but with KluctlDeployments as dependency objects.

## Kubeconfigs and RBAC

As Kluctl is meant to be a CLI-first tool, it expects a kubeconfig to be present while deployments are
performed. The controller will generate such kubeconfigs on-the-fly before performing the actual deployment.

The kubeconfig can be generated from 3 different sources:
1. The default impersonation service account specified at controller startup (via `--default-service-account`)
2. The service account specified via `spec.serviceAccountName` in the KluctlDeployment
3. The secret specified via `spec.kubeConfig` in the KluctlDeployment.

The behavior/functionality of 1. and 2. is comparable to how the [kustomize-controller](https://fluxcd.io/docs/components/kustomize/kustomization/#role-based-access-control)
handles impersonation, with the difference that a kubeconfig with a "default" context is created in-between.

`spec.kubeConfig` will simply load the kubeconfig from `data.value` of the specified secret.

Kluctl [targets](https://kluctl.io/docs/reference/kluctl-project/targets/) specify a context name that is expected to
be present in the kubeconfig while deploying. As the context found in the generated kubeconfig does not necessarily
have the correct name, `spec.renameContexts` allows to rename contexts to the desired names. This is especially useful
when using service account based kubeconfigs, as these always have the same context with the name "default".

Here is an example of a deployment that uses the service account "prod-service-account" and renames the context
appropriately (assuming the Kluctl cluster config for the given target expects a "prod" context):

```yaml
apiVersion: flux.kluctl.io/v1alpha1
kind: KluctlDeployment
metadata:
  name: example
  namespace: flux-system
spec:
  interval: 10m
  sourceRef:
    kind: GitRepository
    name: example
  target: prod
  serviceAccountName: prod-service-account
  renameContexts:
    - oldContext: default
      newContext: prod
```

## Status

When the controller completes a deployments, it reports the result in the `status` sub-resource.

A successful reconciliation sets the ready condition to `true` and updates the revision field:

```yaml
status:
  conditions:
  - lastTransitionTime: "2022-05-04T10:08:39Z"
    message: 'Deployed revision: main/b285d08164011fb642072bc9e3c62c898eba96f5'
    reason: ReconciliationSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-05-04T10:08:39Z"
    message: ReconciliationSucceeded
    reason: ReconciliationSucceeded
    status: "True"
    type: Healthy
  lastAttemptedReconcile:
    deployResult:
      newObjects:
      - id: ms-demo-test__Namespace
        v: v1
      ...
    pruneResult: {}
    revision: main/b285d08164011fb642072bc9e3c62c898eba96f5
    targetHash: 0669d6dbc5be975f90a685bebcf83bc6049f6cf48538c78a7b3862621b8015df
    targetName: test
    time: "2022-05-04T10:08:39Z"
  lastForceReconcileHash: acad6b40f8556cf0b7752d0286e1b45d2855e3c4ba38cb4f515e03ac62236cc0
  lastSuccessfulReconcile:
    deployResult:
      newObjects:
      - id: ms-demo-test__Namespace
        v: v1
      ...
    pruneResult: {}
    revision: main/b285d08164011fb642072bc9e3c62c898eba96f5
    targetHash: 0669d6dbc5be975f90a685bebcf83bc6049f6cf48538c78a7b3862621b8015df
    targetName: test
    time: "2022-05-04T10:08:39Z"
  observedGeneration: 1
```

You can wait for the controller to complete a reconciliation with:

```bash
kubectl wait kluctldeployment/backend --for=condition=ready
```

A failed reconciliation sets the ready condition to `false`:

```yaml
status:
  conditions:
  - lastTransitionTime: "2022-05-04T10:18:11Z"
    message: target invalid-name not found in kluctl project
    reason: PrepareFailed
    status: "False"
    type: Ready
  lastAttemptedReconcile:
    revision: main/b285d08164011fb642072bc9e3c62c898eba96f5
    targetName: invalid-name
    time: "2022-05-04T10:18:11Z"
``` 

> **Note** that the lastSuccessfulReconcile is updated only on a successful reconciliation.
