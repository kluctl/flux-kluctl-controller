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
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

const (
	KluctlDeploymentKind      = "KluctlDeployment"
	KluctlDeploymentFinalizer = "finalizers.flux.kluctl.io"
	MaxConditionMessageLength = 20000
	DisabledValue             = "disabled"
	MergeValue                = "merge"

	KluctlDeployModeFull   = "full-deploy"
	KluctlDeployPokeImages = "poke-images"

	KluctlDeployRequestAnnotation = "deploy.flux.kluctl.io/requestedAt"
)

type KluctlDeploymentSpec struct {
	// Path to the directory containing the .kluctl.yaml file, or the
	// Defaults to 'None', which translates to the root path of the SourceRef.
	// +optional
	Path string `json:"path,omitempty"`

	// Reference of the source where the kluctl project is.
	// The authentication secrets from the source are also used to authenticate
	// dependent git repositories which are cloned while deploying the kluctl project.
	// +required
	SourceRef meta.NamespacedObjectKindReference `json:"sourceRef"`

	// The interval at which to reconcile the KluctlDeployment.
	// By default, the controller will re-deploy and validate the deployment on each reconciliation.
	// To override this behavior, change the DeployInterval and/or ValidateInterval values.
	// +required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	Interval metav1.Duration `json:"interval"`

	// The interval at which to retry a previously failed reconciliation.
	// When not specified, the controller uses the Interval
	// value to retry failures.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`

	// DeployInterval specifies the interval at which to deploy the KluctlDeployment.
	// It defaults to the Interval value, meaning that it will re-deploy on every reconciliation.
	// If you set DeployInterval to a different value,
	// +optional
	DeployInterval *DurationOrNever `json:"deployInterval,omitempty"`

	// DeployOnChanges will cause a re-deployment whenever the rendered resources change in the deployment.
	// This check is performed on every reconciliation. This means that a deployment will be triggered even before
	// the DeployInterval has passed in case something has changed in the rendered resources.
	// +optional
	// +kubebuilder:default:=true
	DeployOnChanges bool `json:"deployOnChanges"`

	// ValidateInterval specifies the interval at which to validate the KluctlDeployment.
	// Validation is performed the same way as with 'kluctl validate -t <target>'.
	// Defaults to the same value as specified in Interval.
	// Validate is also performed whenever a deployment is performed, independent of the value of ValidateInterval
	// +optional
	ValidateInterval *DurationOrNever `json:"validateInterval,omitempty"`

	// Timeout for all operations.
	// Defaults to 'Interval' duration.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ms|s|m|h))+$"
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// This flag tells the controller to suspend subsequent kluctl executions,
	// it does not apply to already started executions. Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// RegistrySecrets is a list of secret references to be used for image registry authentication.
	// The secrets must either have ".dockerconfigjson" included or "registry", "username" and "password".
	// Additionally, "caFile" and "insecure" can be specified.
	// +optional
	RegistrySecrets []meta.LocalObjectReference `json:"registrySecrets,omitempty"`

	// The name of the Kubernetes service account to use while deploying.
	// If not specified, the default service account is used.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// The KubeConfig for deploying to the target cluster.
	// Specifies the kubeconfig to be used when invoking kluctl. Contexts in this kubeconfig must match
	// the context found in the kluctl target. As an alternative, RenameContexts can be used to fix
	// non-matching context names.
	// +optional
	KubeConfig *KubeConfig `json:"kubeConfig"`

	// RenameContexts specifies a list of context rename operations.
	// This is useful when the kluctl target's context does not match with the
	// contexts found in the kubeconfig while deploying. This is the case when using kubeconfigs generated from
	// service accounts, in which case the context name is always "default".
	// +optional
	RenameContexts []RenameContext `json:"renameContexts,omitempty"`

	// Target specifies the kluctl target to deploy
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +required
	Target string `json:"target"`

	// Args specifies dynamic target args.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Args runtime.RawExtension `json:"args,omitempty"`

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

	// DeployMode specifies what deploy mode should be used
	// +kubebuilder:default:=full-deploy
	// +kubebuilder:validation:Enum=full-deploy;poke-images
	// +optional
	DeployMode string `json:"deployMode,omitempty"`

	// Validate enables validation after deploying
	// +kubebuilder:default:=true
	// +optional
	Validate bool `json:"validate"`

	// Prune enables pruning after deploying.
	// +kubebuilder:default:=false
	// +optional
	Prune bool `json:"prune,omitempty"`
}

// GetRetryInterval returns the retry interval
func (in KluctlDeploymentSpec) GetRetryInterval() time.Duration {
	if in.RetryInterval != nil {
		return in.RetryInterval.Duration
	}
	return in.Interval.Duration
}

// KubeConfig references a Kubernetes secret that contains a kubeconfig file.
type KubeConfig struct {
	// SecretRef holds the name of a secret that contains a key with
	// the kubeconfig file as the value. If no key is set, the key will default
	// to 'value'. The secret must be in the same namespace as
	// the Kustomization.
	// It is recommended that the kubeconfig is self-contained, and the secret
	// is regularly updated if credentials such as a cloud-access-token expire.
	// Cloud specific `cmd-path` auth helpers will not function without adding
	// binaries and credentials to the Pod that is responsible for reconciling
	// the KluctlDeployment.
	// +required
	SecretRef meta.SecretKeyReference `json:"secretRef,omitempty"`
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

	// +optional
	LastHandledDeployAt string `json:"lastHandledDeployAt,omitempty"`

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// LastDeployResult is the result of the last deploy command
	// +optional
	LastDeployResult *LastCommandResult `json:"lastDeployResult,omitempty"`

	// LastDeployResult is the result of the last prune command
	// +optional
	LastPruneResult *LastCommandResult `json:"lastPruneResult,omitempty"`

	// LastValidateResult is the result of the last validate command
	// +optional
	LastValidateResult *LastValidateResult `json:"lastValidateResult,omitempty"`

	// CommonLabels are the commonLabels found in the deployment project when the last deployment was done.
	// This is used to perform cleanup/deletion in case the KluctlDeployment project is deleted
	// +optional
	CommonLabels map[string]string `json:"commonLabels,omitempty"`

	// +optional
	RawTarget *string `json:"rawTarget,omitempty"`
}

type ReconcileResultBase struct {
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

	// ObjectsHash is the hash of all rendered objects
	// +optional
	ObjectsHash string `json:"objectsHash,omitempty"`
}

type LastCommandResult struct {
	ReconcileResultBase `json:",inline"`

	// +optional
	RawResult *string `json:"rawResult,omitempty"`

	// +optional
	Error string `json:"error,omitempty"`
}

type LastValidateResult struct {
	ReconcileResultBase `json:",inline"`

	// +optional
	RawResult *string `json:"rawResult,omitempty"`

	// +optional
	Error string `json:"error"`
}

func (r *LastCommandResult) ParseResult() *types.CommandResult {
	if r == nil || r.RawResult == nil {
		return nil
	}

	var ret types.CommandResult
	err := yaml.ReadYamlString(*r.RawResult, &ret)
	if err != nil {
		return nil
	}
	return &ret
}

func (r *LastValidateResult) ParseResult() *types.ValidateResult {
	if r == nil || r.RawResult == nil {
		return nil
	}

	var ret types.ValidateResult
	err := yaml.ReadYamlString(*r.RawResult, &ret)
	if err != nil {
		return nil
	}
	return &ret
}

func (d *KluctlDeploymentStatus) SetRawTarget(target *types.Target) {
	y, err := yaml.WriteYamlString(target)
	if err == nil {
		d.RawTarget = &y
	}
}

func (d *KluctlDeploymentStatus) ParseRawTarget() *types.Target {
	if d.RawTarget == nil {
		return nil
	}
	var ret types.Target
	err := yaml.ReadYamlString(*d.RawTarget, &ret)
	if err != nil {
		return nil
	}
	return &ret
}

func SetDeployResult(k *KluctlDeployment, revision string, result *types.CommandResult, objectHash string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	k.Status.LastDeployResult = &LastCommandResult{
		ReconcileResultBase: ReconcileResultBase{
			AttemptedAt: metav1.Now(),
			Revision:    revision,
			TargetName:  k.Spec.Target,
			ObjectsHash: objectHash,
		},
		Error: errStr,
	}
	if result != nil {
		raw, err := yaml.WriteYamlString(result)
		if err == nil {
			k.Status.LastDeployResult.RawResult = &raw
		}
	}
}

func SetPruneResult(k *KluctlDeployment, revision string, result *types.CommandResult, objectHash string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	k.Status.LastPruneResult = &LastCommandResult{
		ReconcileResultBase: ReconcileResultBase{
			AttemptedAt: metav1.Now(),
			Revision:    revision,
			TargetName:  k.Spec.Target,
			ObjectsHash: objectHash,
		},
		Error: errStr,
	}
	if result != nil {
		raw, err := yaml.WriteYamlString(result)
		if err == nil {
			k.Status.LastPruneResult.RawResult = &raw
		}
	}
}

func SetValidateResult(k *KluctlDeployment, revision string, result *types.ValidateResult, objectHash string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	k.Status.LastValidateResult = &LastValidateResult{
		ReconcileResultBase: ReconcileResultBase{
			AttemptedAt: metav1.Now(),
			Revision:    revision,
			TargetName:  k.Spec.Target,
			ObjectsHash: objectHash,
		},
		Error: errStr,
	}
	if result != nil {
		raw, err := yaml.WriteYamlString(result)
		if err == nil {
			k.Status.LastValidateResult.RawResult = &raw
		}
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Deployed",type="date",JSONPath=".status.lastDeployResult.time",description=""
//+kubebuilder:printcolumn:name="Pruned",type="date",JSONPath=".status.lastPruneResult.time",description=""
//+kubebuilder:printcolumn:name="Validated",type="date",JSONPath=".status.lastValidateResult.time",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// KluctlDeployment is the Schema for the kluctldeployments API
type KluctlDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KluctlDeploymentSpec   `json:"spec,omitempty"`
	Status KluctlDeploymentStatus `json:"status,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (in *KluctlDeployment) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *KluctlDeployment) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
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
