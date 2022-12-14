package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

type GitRef struct {
	// Branch to filter for. Can also be a regex.
	// +optional
	Branch string `json:"branch,omitempty"`

	// Branch to filter for. Can also be a regex.
	// +optional
	Tag string `json:"tag,omitempty"`

	// TODO
	// Commit SHA to check out, takes precedence over all reference fields.
	// +optional
	// Commit string `json:"commit,omitempty"`
}

func (r *GitRef) String() string {
	if r == nil {
		return ""
	}
	if r.Tag != "" {
		return fmt.Sprintf("refs/tags/%s", r.Tag)
	} else if r.Branch != "" {
		return fmt.Sprintf("refs/heads/%s", r.Branch)
	} else {
		return ""
	}
}
