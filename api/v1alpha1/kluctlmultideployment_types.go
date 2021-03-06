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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KluctlMultiDeploymentSpec defines the desired state of KluctlMultiDeployment
type KluctlMultiDeploymentSpec struct {
	KluctlProjectSpec `json:",inline"`
	KluctlTimingSpec  `json:",inline"`

	// TargetPattern is the regex pattern used to match targets
	// +required
	TargetPattern string `json:"targetPattern"`

	// Template is the object template used to create KluctlDeploymet objects
	// +required
	Template KluctlMultiDeploymentTemplate `json:"template"`
}

// KluctlMultiDeploymentTemplate is the template used to create KluctlDeployment objects
type KluctlMultiDeploymentTemplate struct {
	metav1.ObjectMeta `json:",inline"`

	// Spec is the KluctlDeployment spec to be used as a template
	// +required
	Spec KluctlDeploymentTemplateSpec `json:"spec"`
}

// KluctlMultiDeploymentStatus defines the observed state of KluctlMultiDeployment
type KluctlMultiDeploymentStatus struct {
	KluctlProjectStatus `json:",inline"`

	// TargetCount is the number of targets detected
	// +optional
	TargetCount int `json:"targetCount,omitempty"`

	// Targets is the list of detected targets
	// +optional
	Targets []KluctlMultiDeploymentTargetStatus `json:"targets,omitempty"`
}

// KluctlMultiDeploymentTargetStatus describes the status of a single target
type KluctlMultiDeploymentTargetStatus struct {
	// Name is the name of the detected target
	Name string `json:"name"`

	// KluctlDeploymentName is the name of the generated KluctlDeployment object
	KluctlDeploymentName string `json:"kluctlDeploymentName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Pattern",type="string",JSONPath=".spec.targetPattern",description=""
//+kubebuilder:printcolumn:name="Targets",type="integer",JSONPath=".status.targetCount",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// KluctlMultiDeployment is the Schema for the kluctlmultideployments API
type KluctlMultiDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KluctlMultiDeploymentSpec   `json:"spec,omitempty"`
	Status KluctlMultiDeploymentStatus `json:"status,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (in *KluctlMultiDeployment) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *KluctlMultiDeployment) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetStatusConditions returns a pointer to the Status.Conditions slice.
// Deprecated: use GetConditions instead.
func (in *KluctlMultiDeployment) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *KluctlMultiDeployment) GetDependsOn() []meta.NamespacedObjectReference {
	return in.Spec.DependsOn
}

func (in *KluctlMultiDeployment) GetKluctlProject() *KluctlProjectSpec {
	return &in.Spec.KluctlProjectSpec
}

func (in *KluctlMultiDeployment) GetKluctlTiming() *KluctlTimingSpec {
	return &in.Spec.KluctlTimingSpec
}

func (in *KluctlMultiDeployment) GetKluctlStatus() *KluctlProjectStatus {
	return &in.Status.KluctlProjectStatus
}

func (in *KluctlMultiDeployment) GetFullStatus() any {
	return &in.Status
}

func (in *KluctlMultiDeployment) SetFullStatus(s any) {
	s2, ok := s.(*KluctlMultiDeploymentStatus)
	if !ok {
		panic("not a KluctlMultiDeploymentStatus")
	}
	in.Status = *s2
}

//+kubebuilder:object:root=true

// KluctlMultiDeploymentList contains a list of KluctlMultiDeployment
type KluctlMultiDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KluctlMultiDeployment `json:"items"`
}

func (in *KluctlMultiDeploymentList) GetItems() []client.Object {
	var ret []client.Object
	for _, x := range in.Items {
		x := x
		ret = append(ret, &x)
	}

	return ret
}

func init() {
	SchemeBuilder.Register(&KluctlMultiDeployment{}, &KluctlMultiDeploymentList{})
}
