/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/fluxcd/pkg/apis/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"time"
)

const (
	KluctlDeploymentKind      = "KluctlDeployment"
	KluctlDeploymentFinalizer = "finalizers.flux.kluctl.io"
	MaxConditionMessageLength = 20000
	DisabledValue             = "disabled"
	MergeValue                = "merge"
)

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

	// The KubeConfig for deploying to the target cluster.
	// Specifies the kubeconfig to be used when invoking kluctl. Contexts in this kubeconfig must match
	// the cluster config found in the kluctl project. As alternative, RenameContexts can be used to fix
	// non-matching context names. If KubeConfig is not specified, the service account of the controller is
	// used to generate a kubeconfig.
	// +optional
	KubeConfig *KubeConfig `json:"kubeConfig"`

	// RenameContexts specifies a list of context rename operations.
	// This is useful when the kluctl project's cluster configs specify contexts that do not match with the
	// contexts found in the kubeconfig while deploying. This is the case when re-using the in-cluster kubeconfig
	// of the controller which always has the name "default"
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

// RenameContext specifies a single rename of a context
type RenameContext struct {
	// OldContext is the name of the context to be renamed
	// +required
	OldContext string `json:"oldContext"`

	// NewContext is the new name of the context
	// +required
	NewContext string `json:"newContext"`
}

// KluctlDeploymentStatus defines the observed state of KluctlDeployment
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

// InvolvedRepo represents a git repository and all involved refs
type InvolvedRepo struct {
	// URL is the url of the involved git repository
	URL string `json:"url"`

	// Patterns is a list of pattern+refs combinations
	Patterns []InvolvedRepoPattern `json:"patterns"`
}

// InvolvedRepoPattern represents a ref pattern and the found refs
type InvolvedRepoPattern struct {
	// Pattern is a regex to filter refs
	Pattern string `json:"pattern"`

	// Refs is the filtered list of refs
	Refs map[string]string `json:"refs"`
}

// ReconcileAttempt describes an attempt to reconcile
type ReconcileAttempt struct {
	// AttemptedAt is the time when the attempt was performed
	// +required
	AttemptedAt metav1.Time `json:"time"`

	// Revision is the source revision. Please note that kluctl projects have
	// dependent git repositories which are not considered in the source revision
	// +optional
	Revision string `json:"revision,omitempty"`

	// TargetName is the name of the target
	// +required
	TargetName string `json:"targetName"`

	// TargetHash is the hash of the target configuration
	// +optional
	TargetHash string `json:"targetHash,omitempty"`

	// DeployResult is the command result of the deploy command
	// +optional
	DeployResult *CommandResult `json:"deployResult,omitempty"`

	// PruneResult is the command result of the prune command
	// +optional
	PruneResult *CommandResult `json:"pruneResult,omitempty"`
}

// KluctlDeploymentProgressing resets the conditions of the given KluctlDeployment to a single
// ReadyCondition with status ConditionUnknown.
func KluctlDeploymentProgressing(k KluctlDeployment, message string) KluctlDeployment {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  metav1.ConditionUnknown,
		Reason:  meta.ProgressingReason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(k.GetStatusConditions(), newCondition)
	return k
}

// SetKluctlDeploymentHealthiness sets the HealthyCondition status for a KluctlDeployment.
func SetKluctlDeploymentHealthiness(k *KluctlDeployment, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    HealthyCondition,
		Status:  status,
		Reason:  reason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(k.GetStatusConditions(), newCondition)
}

// SetKluctlDeploymentReadiness sets the ReadyCondition, ObservedGeneration, and LastAttemptedReconcile, on the KluctlDeployment.
func SetKluctlDeploymentReadiness(k *KluctlDeployment, status metav1.ConditionStatus, reason, message string, revision string, targetHash string, deployResult *CommandResult, pruneResult *CommandResult) {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  status,
		Reason:  reason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(k.GetStatusConditions(), newCondition)

	k.Status.ObservedGeneration = k.Generation
	k.Status.LastAttemptedReconcile = &ReconcileAttempt{
		AttemptedAt:  metav1.Now(),
		Revision:     revision,
		TargetName:   k.Spec.Target,
		TargetHash:   targetHash,
		DeployResult: deployResult,
		PruneResult:  pruneResult,
	}
}

// KluctlDeploymentNotReady registers a failed apply attempt of the given KluctlDeployment.
func KluctlDeploymentNotReady(k KluctlDeployment, revision string, targetHash string, deployResult *CommandResult, pruneResult *CommandResult, reason, message string) KluctlDeployment {
	SetKluctlDeploymentReadiness(&k, metav1.ConditionFalse, reason, trimString(message, MaxConditionMessageLength), revision, targetHash, deployResult, pruneResult)
	return k
}

// KluctlDeploymentReady registers a successful deploy attempt of the given KluctlDeployment.
func KluctlDeploymentReady(k KluctlDeployment, revision string, targetHash string, deployResult *CommandResult, pruneResult *CommandResult, reason, message string) KluctlDeployment {
	SetKluctlDeploymentReadiness(&k, metav1.ConditionTrue, reason, trimString(message, MaxConditionMessageLength), revision, targetHash, deployResult, pruneResult)
	SetKluctlDeploymentHealthiness(&k, metav1.ConditionTrue, reason, reason)
	k.Status.LastSuccessfulReconcile = k.Status.LastAttemptedReconcile.DeepCopy()
	return k
}

// CalcForceReconcileHash calculates a hash from all values of the spec that must cause a forced reconciliation
func (in KluctlDeployment) CalcForceReconcileHash(scheme *runtime.Scheme) (string, error) {
	onlySpec := KluctlDeployment{
		Spec: in.Spec,
	}

	h := sha256.New()
	s := json.NewSerializerWithOptions(nil, scheme, scheme, json.SerializerOptions{})
	err := s.Encode(&onlySpec, h)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)[:]), nil
}

// GetTimeout returns the timeout with default.
func (in KluctlDeployment) GetTimeout() time.Duration {
	duration := in.Spec.Interval.Duration - 30*time.Second
	if in.Spec.Timeout != nil {
		duration = in.Spec.Timeout.Duration
	}
	if duration < 30*time.Second {
		return 30 * time.Second
	}
	return duration
}

// GetRetryInterval returns the retry interval
func (in KluctlDeployment) GetRetryInterval() time.Duration {
	if in.Spec.RetryInterval != nil {
		return in.Spec.RetryInterval.Duration
	}
	return in.GetRequeueAfter()
}

// GetRequeueAfter returns the duration after which the KluctlDeployment must be
// reconciled again.
func (in KluctlDeployment) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetDependsOn returns the list of dependencies across-namespaces.
func (in KluctlDeployment) GetDependsOn() []meta.NamespacedObjectReference {
	return in.Spec.DependsOn
}

// GetConditions returns the status conditions of the object.
func (in KluctlDeployment) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *KluctlDeployment) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetStatusConditions returns a pointer to the Status.Conditions slice.
// Deprecated: use GetConditions instead.
func (in *KluctlDeployment) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KluctlDeployment is the Schema for the kluctldeployments API
type KluctlDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KluctlDeploymentSpec   `json:"spec,omitempty"`
	Status KluctlDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KluctlDeploymentList contains a list of KluctlDeployment
type KluctlDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KluctlDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KluctlDeployment{}, &KluctlDeploymentList{})
}

func trimString(str string, limit int) string {
	if len(str) <= limit {
		return str
	}

	return str[0:limit] + "..."
}
