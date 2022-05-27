package v1alpha1

import (
	"github.com/fluxcd/pkg/apis/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type KluctlProjectSpec struct {
	// DependsOn may contain a meta.NamespacedObjectReference slice
	// with references to resources that must be ready before this
	// kluctl project can be deployed.
	// +optional
	DependsOn []meta.NamespacedObjectReference `json:"dependsOn,omitempty"`

	// Path to the directory containing the .kluctl.yaml file, or the
	// Defaults to 'None', which translates to the root path of the SourceRef.
	// +optional
	Path string `json:"path,omitempty"`

	// Reference of the source where the kluctl project is.
	// The authentication secrets from the source are also used to authenticate
	// dependent git repositories which are cloned while deploying the kluctl project.
	// +required
	SourceRef CrossNamespaceSourceReference `json:"sourceRef"`

	// This flag tells the controller to suspend subsequent kluctl executions,
	// it does not apply to already started executions. Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

type KluctlTimingSpec struct {
	// The interval at which to reconcile the KluctlDeployment.
	// +required
	Interval metav1.Duration `json:"interval"`

	// The interval at which to retry a previously failed reconciliation.
	// When not specified, the controller uses the KluctlDeploymentSpec.Interval
	// value to retry failures.
	// +optional
	RetryInterval *metav1.Duration `json:"retryInterval,omitempty"`

	// Timeout for all operations.
	// Defaults to 'Interval' duration.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// KluctlProjectStatus defines the observed state of KluctlProjectStatus
type KluctlProjectStatus struct {
	meta.ReconcileRequestStatus `json:",inline"`

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`
}

// GetDependsOn returns the list of dependencies across-namespaces.
func (in KluctlProjectSpec) GetDependsOn() []meta.NamespacedObjectReference {
	return in.DependsOn
}

// GetTimeout returns the timeout with default.
func (in KluctlTimingSpec) GetTimeout() time.Duration {
	duration := in.Interval.Duration - 30*time.Second
	if in.Timeout != nil {
		duration = in.Timeout.Duration
	}
	if duration < 30*time.Second {
		return 30 * time.Second
	}
	return duration
}

// GetRetryInterval returns the retry interval
func (in KluctlTimingSpec) GetRetryInterval() time.Duration {
	if in.RetryInterval != nil {
		return in.RetryInterval.Duration
	}
	return in.GetRequeueAfter()
}

// GetRequeueAfter returns the duration after which the KluctlDeployment must be
// reconciled again.
func (in KluctlTimingSpec) GetRequeueAfter() time.Duration {
	return in.Interval.Duration
}

// KluctlProjectProgressing resets the conditions of the given KluctlProjectStatus to a single
// ReadyCondition with status ConditionUnknown.
func KluctlProjectProgressing(k *KluctlProjectStatus, message string) {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  metav1.ConditionUnknown,
		Reason:  meta.ProgressingReason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(&k.Conditions, newCondition)
}

// SetKluctlProjectHealthiness sets the HealthyCondition status for a KluctlProjectStatus.
func SetKluctlProjectHealthiness(k *KluctlProjectStatus, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    HealthyCondition,
		Status:  status,
		Reason:  reason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(&k.Conditions, newCondition)
}

// SetKluctlProjectReadiness sets the ReadyCondition, ObservedGeneration, and LastAttemptedReconcile, on the KluctlProjectStatus.
func SetKluctlProjectReadiness(k *KluctlProjectStatus, status metav1.ConditionStatus, reason, message string, generation int64, revision string) {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  status,
		Reason:  reason,
		Message: trimString(message, MaxConditionMessageLength),
	}
	apimeta.SetStatusCondition(&k.Conditions, newCondition)

	k.ObservedGeneration = generation
	k.LastAttemptedRevision = revision
}
