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

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern="^(([0-9]+(\\.[0-9]+)?(ms|s|m|h))+)|never$"
type DurationOrNever struct {
	Duration metav1.Duration
	Never    bool
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (d *DurationOrNever) UnmarshalJSON(b []byte) error {
	if string(b) == `"never"` {
		d.Never = true
		d.Duration.Reset()
		return nil
	} else {
		d.Never = false
		return d.Duration.UnmarshalJSON(b)
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (d DurationOrNever) MarshalJSON() ([]byte, error) {
	if d.Never {
		return []byte(`"never"`), nil
	}
	return d.Duration.MarshalJSON()
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (d DurationOrNever) ToUnstructured() interface{} {
	if d.Never {
		return "never"
	}
	return d.Duration.ToUnstructured()
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
//
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (_ DurationOrNever) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (_ DurationOrNever) OpenAPISchemaFormat() string { return "" }

type KluctlTimingSpec struct {
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

// GetRetryInterval returns the retry interval
func (in KluctlTimingSpec) GetRetryInterval() time.Duration {
	if in.RetryInterval != nil {
		return in.RetryInterval.Duration
	}
	return in.Interval.Duration
}
