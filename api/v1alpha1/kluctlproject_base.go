package v1alpha1

import (
	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type KluctlProjectSpec struct {
	// Path to the directory containing the .kluctl.yaml file, or the
	// Defaults to 'None', which translates to the root path of the SourceRef.
	// +optional
	Path string `json:"path,omitempty"`

	// Reference of the source where the kluctl project is.
	// The authentication secrets from the source are also used to authenticate
	// dependent git repositories which are cloned while deploying the kluctl project.
	// +required
	SourceRef meta.NamespacedObjectKindReference `json:"sourceRef"`
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

	// DeployInterval specifies the interval at which to deploy the KluctlDeployment.
	// This is independent of the 'Interval' value, which only causes deployments if some deployment objects have
	// changed.
	// +optional
	DeployInterval *metav1.Duration `json:"deployInterval"`

	// ValidateInterval specifies the interval at which to validate the KluctlDeployment.
	// Validation is performed the same way as with 'kluctl validate -t <target>'.
	// Defaults to 5m.
	// +kubebuilder:default:="5m"
	// +optional
	ValidateInterval metav1.Duration `json:"validateInterval"`

	// Timeout for all operations.
	// Defaults to 'Interval' duration.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// This flag tells the controller to suspend subsequent kluctl executions,
	// it does not apply to already started executions. Defaults to false.
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// KluctlProjectStatus defines the observed state of KluctlProjectStatus
type KluctlProjectStatus struct {
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
