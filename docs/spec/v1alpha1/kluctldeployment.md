# KluctlDeployment

The `KluctlDeployment` API defines a deployment of a [target](https://kluctl.io/docs/reference/kluctl-project/targets/)
from a [Kluctl Project](https://kluctl.io/docs/reference/kluctl-project/).

## Specification

A **KluctlDeployment** object defines the source of the root project by referencing an object 
managed by [source-controller](https://github.com/fluxcd/source-controller),
the path to the .kluctl.yaml config file within that source,
and the interval at which the target is deployed to the cluster.

```go
// KluctlDeploymentSpec defines the desired state of KluctlDeployment
type KluctlDeploymentSpec struct {
	// DependsOn may contain a meta.NamespacedObjectReference slice
	// with references to resources that must be ready before this
	// kluctl project can be deployed.
	// +optional
	DependsOn []meta.NamespacedObjectReference `json:"dependsOn,omitempty"`

	// The interval at which to reconcile the KluctlDeployment.
	// +required
	Interval metav1.Duration `json:"interval"`

	// The interval at which to retry a previously failed reconciliation.
	// When not specified, the controller uses the KluctlDeploymentSpec.Interval
	// value to retry failures.
	// +optional
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`

	// Path to the directory containing the .kluctl.yaml file, or the
	// Defaults to 'None', which translates to the root path of the SourceRef.
	// +optional
	Path string `json:"path,omitempty"`

	// Reference of the source where the kluctl project is.
	// The authentication secrets from the source are also used to authenticate
	// dependent git repositories which are cloned while deploying the kluctl project.
	// +required
	SourceRef CrossNamespaceSourceReference `json:"sourceRef"`

	// RegistrySecrets is a list of secret references to be used for image registry authentication.
	// The secrets must either have ".dockerconfigjson" included or "registry", "username" and "password".
	// Additionally, "caFile" and "insecure" can be specified.
	// +optional
	RegistrySecrets []meta.LocalObjectReference `json:"registrySecrets,omitempty"`

	// This flag tells the controller to suspend subsequent kluctl executions,
	// it does not apply to already started executions. Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// Timeout for all operations.
	// Defaults to 'Interval' duration.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Target specifies the kluctl target to deploy
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +required
	Target string `json:"target"`

	// The name of the Kubernetes service account to use while deploying.
	// If not specified, the default service account is used.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// The KubeConfig for deploying to the target cluster.
	// Specifies the kubeconfig to be used when invoking kluctl. Contexts in this kubeconfig must match
	// the cluster config found in the kluctl project. As alternative, RenameContexts can be used to fix
	// non-matching context names.
	// +optional
	KubeConfig *KubeConfig `json:"kubeConfig"`

	// RenameContexts specifies a list of context rename operations.
	// This is useful when the kluctl project's cluster configs specify contexts that do not match with the
	// contexts found in the kubeconfig while deploying. This is the case when using kubeconfigs generated from
	// service accounts, in which case the context name is always "default".
	// +optional
	RenameContexts []RenameContext `json:"renameContexts,omitempty"`

	// Args specifies dynamic target args.
	// Only arguments defined by 'dynamicArgs' of the target are allowed.
	// +optional
	Args map[string]string `json:"args,omitempty"`

	// UpdateImages instructs kluctl to update dynamic images.
	// Equivalent to using '-u' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	UpdateImages bool `json:"updateImages,omitempty"`

	// Images contains a list of fixed image overrides.
	// Equivalent to using '--fixed-images-file' when calling kluctl.
	// +optional
	Images []FixedImage `json:"images,omitempty"`

	// DryRun instructs kluctl to run everything in dry-run mode.
	// Equivalent to using '--dry-run' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// NoWait instructs kluctl to not wait for any resources to become ready, including hooks.
	// Equivalent to using '--no-wait' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	NoWait bool `json:"noWait,omitempty"`

	// ForceApply instructs kluctl to force-apply in case of SSA conflicts.
	// Equivalent to using '--force-apply' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	ForceApply bool `json:"forceApply,omitempty"`

	// ReplaceOnError instructs kluctl to replace resources on error.
	// Equivalent to using '--replace-on-error' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	ReplaceOnError bool `json:"replaceOnError,omitempty"`

	// ForceReplaceOnError instructs kluctl to force-replace resources in case a normal replace fails.
	// Equivalent to using '--force-replace-on-error' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	ForceReplaceOnError bool `json:"forceReplaceOnError,omitempty"`

	// ForceReplaceOnError instructs kluctl to abort deployments immediately when something fails.
	// Equivalent to using '--abort-on-error' when calling kluctl.
	// +kubebuilder:default:=false
	// +optional
	AbortOnError bool `json:"abortOnError,omitempty"`

	// IncludeTags instructs kluctl to only include deployments with given tags.
	// Equivalent to using '--include-tag' when calling kluctl.
	// +optional
	IncludeTags []string `json:"includeTags,omitempty"`

	// ExcludeTags instructs kluctl to exclude deployments with given tags.
	// Equivalent to using '--exclude-tag' when calling kluctl.
	// +optional
	ExcludeTags []string `json:"excludeTags,omitempty"`

	// IncludeDeploymentDirs instructs kluctl to only include deployments with the given dir.
	// Equivalent to using '--include-deployment-dir' when calling kluctl.
	// +optional
	IncludeDeploymentDirs []string `json:"includeDeploymentDirs,omitempty"`

	// ExcludeDeploymentDirs instructs kluctl to exclude deployments with the given dir.
	// Equivalent to using '--exclude-deployment-dir' when calling kluctl.
	// +optional
	ExcludeDeploymentDirs []string `json:"excludeDeploymentDirs,omitempty"`

	// Prune enables pruning after deploying.
	// +kubebuilder:default:=false
	// +optional
	Prune bool `json:"prune,omitempty"`
}
```

KubeConfig references a Kubernetes secret generated by CAPI:

```go
// KubeConfig references a Kubernetes secret that contains a kubeconfig file.
type KubeConfig struct {
	// SecretRef holds the name to a secret that contains a 'value' key with
	// the kubeconfig file as the value. It must be in the same namespace as
	// the KluctlDeployment.
	// It is recommended that the kubeconfig is self-contained, and the secret
	// is regularly updated if credentials such as a cloud-access-token expire.
	// Cloud specific `cmd-path` auth helpers will not function without adding
	// binaries and credentials to the Pod that is responsible for reconciling
	// the KluctlDeployment.
	// +required
	SecretRef meta.LocalObjectReference `json:"secretRef,omitempty"`
}
```

RenameContext specifies a single rename of a context

```go
// RenameContext specifies a single rename of a context
type RenameContext struct {
	// OldContext is the name of the context to be renamed
	// +required
	OldContext string `json:"oldContext"`

	// NewContext is the new name of the context
	// +required
	NewContext string `json:"newContext"`
}
```

The status sub-resource records the result of the last reconciliation:

```go
type KluctlDeploymentStatus struct {
	meta.ReconcileRequestStatus `json:",inline"`

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastForceReconcileHash contains a hash of all values from the spec that must cause a forced
	// reconcile.
	// +optional
	LastForceReconcileHash string `json:"lastForceReconcileHash,omitempty"`

	// InvolvedRepos is a list of repositories and refs involved with this kluctl project
	// +optional
	InvolvedRepos []InvolvedRepo `json:"involvedRepos,omitempty"`

	// The last attempted reconcile.
	// +optional
	LastAttemptedReconcile *ReconcileAttempt `json:"lastAttemptedReconcile,omitempty"`

	// The last successfully reconcile attempt.
	// +optional
	LastSuccessfulReconcile *ReconcileAttempt `json:"lastSuccessfulReconcile,omitempty"`
}
```

Status condition types:

```go
const (
	// ReadyCondition is the name of the condition that
	// records the readiness status of a KluctlDeployment.
	ReadyCondition string = "Ready"
)
```

Status condition reasons:

```go
const (
	// PruneFailedReason represents the fact that the
	// pruning of the KluctlDeployment failed.
	PruneFailedReason string = "PruneFailed"

	// ArtifactFailedReason represents the fact that the
	// source artifact download failed.
	ArtifactFailedReason string = "ArtifactFailed"

	// PrepareFailedReason represents failure in the kluctl preparation phase
	PrepareFailedReason string = "PrepareFailed"

	// DeployFailedReason represents the fact that the
	// kluctl deploy command failed.
	DeployFailedReason string = "DeployFailed"
	
	// DependencyNotReadyReason represents the fact that
	// one of the dependencies is not ready.
	DependencyNotReadyReason string = "DependencyNotReady"

	// ReconciliationSucceededReason represents the fact that
	// the reconciliation succeeded.
	ReconciliationSucceededReason string = "ReconciliationSucceeded"

	// ReconciliationSkippedReason represents the fact that
	// the reconciliation was skipped due to an unchanged target.
	ReconciliationSkippedReason string = "ReconciliationSkipped"
)
```

## Source reference

The KluctlDeployment `spec.sourceRef` is a reference to an object managed by
[source-controller](https://github.com/fluxcd/source-controller). When the source
[revision](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/common.md#source-status) 
changes, it generates a Kubernetes event that triggers a reconciliation attempt.

Source supported types:

* [GitRepository](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/gitrepositories.md)
* [Bucket](https://github.com/fluxcd/source-controller/blob/master/docs/spec/v1alpha1/buckets.md)

The Kluctl project found in the referenced source is also called the "root project". It might
contain references to [external project](https://kluctl.io/docs/reference/kluctl-project/external-projects/),
meaning that other dependent Git projects might get involved without you explicitely defining them via
GitRepository objects. In this case, the controller will re-use the credentials from the root project's
GitRepository for further authentication.

## Reconciliation

The KluctlDeployment `spec.interval` tells the controller at which interval to try reconciliations.
The interval time units are `s`, `m` and `h` e.g. `interval: 5m`, the minimum value should be over 60 seconds.

A reconciliation attempt does not necessarily lead to an actual deployment. The controller keeps track of the
last attempted project revision and only re-deploys in case something has changed. For this, all sources, including
from external projects, are taken into account.

The KluctlDeployment reconciliation can be suspended by setting `spec.susped` to `true`.

The controller can be told to reconcile the KluctlDeployment outside of the specified interval
by annotating the KluctlDeployment object with:

```go
const (
	// ReconcileAtAnnotation is the annotation used for triggering a
	// reconciliation outside of the defined schedule.
	ReconcileAtAnnotation string = "fluxcd.io/reconcileAt"
)
```

On-demand execution example:

```bash
kubectl annotate --overwrite kluctldeployment/podinfo fluxcd.io/reconcileAt="$(date +%s)"
```

This will also cause a forced re-deployment in case the Kluctl project has not changed.

## Pruning

To enable pruning, set `spec.prune` to `true`. This will cause the controller to run `kluctl prune` after each
successful deployment.

## Kubeconfigs and RBAC

As Kluctl is meant to be a CLI-first tool, it expects a kubeconfig to be present while deployments are
performed. The controller will generate such kubeconfigs on-the-fly before performing the actual deployment.

The kubeconfig can be generated from 3 different sources:
1. The default impersonation service account specified at controller startup (via `--default-service-account`)
2. The service account specified via `spec.serviceAccountName` in the KluctlDeployment
3. The secret specified via `spec.kubeConfig` in the KluctlDeployment.

The behavior/functionality of 1. and 2. is comparable to how the [kustomize-controller](https://fluxcd.io/docs/components/kustomize/kustomization/#role-based-access-control)
handles impersonation, withe difference that a kubeconfig with a "default" context is created in-between.

`spec.kubeConfig` will simply load the kubeconfig from `data.value` of the specified secret.

Kluctl [cluster configs](https://kluctl.io/docs/reference/cluster-configs/) specify a context name that is expected to
be present in the kubeconfig while deploying. As the context found in the generated kubeconfig does not necessarly
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
