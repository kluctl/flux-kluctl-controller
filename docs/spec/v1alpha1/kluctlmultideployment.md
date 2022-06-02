# KluctlMultiDeployment

The `KluctlMultiDeployment` API defines a template that is used to create multiple `KluctlDeployment` objects from a
single kluctl project. It specifies a `targetPattern` that is used to match against known targets, for which
`KluctlDeployment` objects are created then.

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
kind: KluctlMultiDeployment
metadata:
  name: microservices-demo-multi
spec:
  interval: 5m
  path: "./microservices-demo/3-templating-and-multi-env/"
  sourceRef:
    kind: GitRepository
    name: microservices-demo

  # Specifies a regex to be matched against the project's targets. A KluctlDeployment will be created for each
  # matching target.
  targetPattern: local|test

  template:
    spec:
      interval: 5m
      timeout: 2m
      prune: true
      # kluctl targets specify the expected context name, which does not necessarily match the context name
      # found while it is deployed via the controller. This means we must pass a kubeconfig to kluctl that has the
      # context renamed to the one that it expects.
      renameContexts:
        - oldContext: default
          newContext: kind-kind
```

In the above example, two objects are being created, a GitRepository that points to the Kluctl project and KluctlMultiDeployment
that defines the target pattern and KluctlDeployment template.

## Reconciliation

Reconciliation of deployments is performed whenever the interval time passes or the source changes. Each matching target will
result in a single KluctlDeployment object which is based on the specified template. The template spec is
identical to the KluctlDeployment spec, but without `dependsOn`, `sourceRef`, `path`, `suspend` and `target`. The first
4 omitted fields are reused from the KluctlMultiDeployment spec. `target` is set to the target name that matched the
targetPattern.

Whenever a target belonging to an already created KluctlDeployment is removed from the project, the controller will
also remove the corresponding KluctlDeployment object. This will in turn cause finalization of the KluctlDeployment
object, which will cause a `kluctl delete` invocation on the target (only if `prune` is set to `true`).

Target matching happens for normal target and for dynamic targets as well, meaning that if targets are loaded from
other git repositories, this is taken into consideration as well.
