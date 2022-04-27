package v1alpha1

import (
	"fmt"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
)

// ResourceRef contains the information necessary to locate a resource within a cluster.
type ResourceRef struct {
	// ID is the string representation of the Kubernetes resource object's metadata,
	// in the format '<namespace>_<name>_<group>_<kind>'.
	ID string `json:"id"`

	// Version is the API version of the Kubernetes resource object's kind.
	Version string `json:"v"`
}

func ConvertResourceRef(ref *k8s.ObjectRef) *ResourceRef {
	if ref == nil {
		return nil
	}
	id := fmt.Sprintf("%s_%s_%s", ref.Name, ref.GVK.Group, ref.GVK.Kind)
	if ref.Namespace != "" {
		id = fmt.Sprintf("%s_%s", id, ref.Namespace)
	}
	return &ResourceRef{ID: id, Version: ref.GVK.Version}
}

type FixedImage struct {
	Image         string       `json:"image"`
	ResultImage   string       `json:"resultImage"`
	DeployedImage *string      `json:"deployedImage,omitempty"`
	RegistryImage *string      `json:"registryImage,omitempty"`
	Namespace     *string      `json:"namespace,omitempty"`
	Object        *ResourceRef `json:"object,omitempty"`
	Deployment    *string      `json:"deployment,omitempty"`
	Container     *string      `json:"container,omitempty"`
	VersionFilter *string      `json:"versionFilter,omitempty"`
	DeployTags    []string     `json:"deployTags,omitempty"`
	DeploymentDir *string      `json:"deploymentDir,omitempty"`
}

func ConvertFixedImage(fi types.FixedImage) *FixedImage {
	return &FixedImage{
		Image:         fi.Image,
		ResultImage:   fi.ResultImage,
		DeployedImage: fi.DeployedImage,
		RegistryImage: fi.RegistryImage,
		Namespace:     fi.Namespace,
		Object:        ConvertResourceRef(fi.Object),
		Deployment:    fi.Deployment,
		Container:     fi.Container,
		VersionFilter: fi.VersionFilter,
		DeployTags:    fi.DeployTags,
		DeploymentDir: fi.DeploymentDir,
	}
}

type Change struct {
	Type        string `json:"type"`
	JsonPath    string `json:"jsonPath"`
	UnifiedDiff string `json:"unifiedDiff,omitempty"`
}

type DeploymentError struct {
	Ref   ResourceRef `json:"ref"`
	Error string      `json:"error"`
}

type CommandResult struct {
	NewObjects     []ResourceRef     `json:"newObjects,omitempty"`
	ChangedObjects []ResourceRef     `json:"changedObjects,omitempty"`
	HookObjects    []ResourceRef     `json:"hookObjects,omitempty"`
	OrphanObjects  []ResourceRef     `json:"orphanObjects,omitempty"`
	DeletedObjects []ResourceRef     `json:"deletedObjects,omitempty"`
	Errors         []DeploymentError `json:"errors,omitempty"`
	Warnings       []DeploymentError `json:"warnings,omitempty"`
	SeenImages     []FixedImage      `json:"seenImages,omitempty"`
}

func ConvertCommandResult(cmdResult *types.CommandResult) *CommandResult {
	if cmdResult == nil {
		return nil
	}
	var ret CommandResult
	for _, x := range cmdResult.NewObjects {
		ret.NewObjects = append(ret.NewObjects, *ConvertResourceRef(&x.Ref))
	}
	for _, x := range cmdResult.ChangedObjects {
		ret.ChangedObjects = append(ret.ChangedObjects, *ConvertResourceRef(&x.Ref))
	}
	for _, x := range cmdResult.HookObjects {
		ret.HookObjects = append(ret.HookObjects, *ConvertResourceRef(&x.Ref))
	}
	for _, x := range cmdResult.OrphanObjects {
		ret.OrphanObjects = append(ret.OrphanObjects, *ConvertResourceRef(&x))
	}
	for _, x := range cmdResult.DeletedObjects {
		ret.DeletedObjects = append(ret.DeletedObjects, *ConvertResourceRef(&x))
	}
	for _, x := range cmdResult.Errors {
		ret.Errors = append(ret.Errors, DeploymentError{Ref: *ConvertResourceRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.Warnings {
		ret.Warnings = append(ret.Warnings, DeploymentError{Ref: *ConvertResourceRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.SeenImages {
		ret.SeenImages = append(ret.SeenImages, *ConvertFixedImage(x))
	}
	return &ret
}

type ValidateResultEntry struct {
	Ref        ResourceRef `json:"ref"`
	Annotation string      `json:"annotation"`
	Message    string      `json:"message"`
}

type ValidateResult struct {
	Ready    bool                  `json:"ready"`
	Warnings []DeploymentError     `json:"warnings,omitempty"`
	Errors   []DeploymentError     `json:"errors,omitempty"`
	Results  []ValidateResultEntry `json:"results,omitempty"`
}
