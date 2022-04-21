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
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/source-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	KluctlProjectKind      = "KluctlProject"
	KluctlProjectFinalizer = "finalizers.kluctl.io"
)

// KluctlProjectSpec defines the desired state of KluctlProject
type KluctlProjectSpec struct {
	// URL specifies the Git repository URL, it can be an HTTP/S or SSH address.
	// +kubebuilder:validation:Pattern="^(http|https|ssh)://"
	// +required
	URL string `json:"url"`

	// Reference specifies the Git reference to resolve and monitor for
	// changes, defaults to the default branch.
	// +optional
	Ref *string `json:"ref,omitempty"`

	// Suspend tells the controller to suspend the reconciliation of this
	// KluctlProject.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// SecretRef specifies the Secret containing authentication credentials for
	// the KluctlProject. The same credentials are used for all involved repos.
	// For HTTPS repositories the Secret must contain 'username' and 'password'
	// fields.
	// For SSH repositories the Secret must contain 'identity', 'identity.pub'
	// and 'known_hosts' fields.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// Interval at which to check the KluctlProject and all involved repos for updates.
	// +required
	Interval metav1.Duration `json:"interval"`

	// Timeout for Git operations like cloning, defaults to 60s.
	// +kubebuilder:default="60s"
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (in KluctlProject) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *KluctlProject) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the GitRepository must be
// reconciled again.
func (in KluctlProject) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

// GetArtifact returns the latest Artifact from the GitRepository if present in
// the status sub-resource.
func (in *KluctlProject) GetArtifact() *v1beta2.Artifact {
	return in.Status.Artifact
}

// KluctlProjectStatus defines the observed state of KluctlProject
type KluctlProjectStatus struct {
	meta.ReconcileRequestStatus `json:",inline"`

	// ObservedGeneration is the last observed generation of the KluctlProject
	// object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the KluctlProject.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// KluctlProjectStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the last successful KluctlProject reconciliation.
	// +optional
	Artifact *v1beta2.Artifact `json:"artifact,omitempty"`

	// ArchiveInfo holds infos about the archive
	// +optional
	ArchiveInfo *ArchiveInfo `json:"archiveInfo,omitempty"`
}

// ArchiveInfo holds infos related to the project archive
type ArchiveInfo struct {
	// ArchiveHash is the hash of archive.tar.gz
	// +required
	ArchiveHash string `json:"archiveHash"`

	// MetdataHash is the hash of metadata.yml
	// +required
	MetadataHash string `json:"metadataHash"`

	// InvolvedRepos is a list of repositories and refs involved with this KluctlProject
	InvolvedRepos []InvolvedRepo `json:"involvedRepos"`

	// Targets is a list of targets found in the KluctlProject
	Targets []TargetInfo `json:"targets"`
}

func (a ArchiveInfo) Revision() string {
	return fmt.Sprintf("%s-%s", a.ArchiveHash, a.MetadataHash)
}

type TargetInfo struct {
	// Name is the name of the target
	// +required
	Name string `json:"name"`

	// TargetHash is the hash of the target configuration
	// +required
	TargetHash string `json:"targetHash"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KluctlProject is the Schema for the kluctlprojects API
type KluctlProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KluctlProjectSpec   `json:"spec,omitempty"`
	Status KluctlProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KluctlProjectList contains a list of KluctlProject
type KluctlProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KluctlProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KluctlProject{}, &KluctlProjectList{})
}
