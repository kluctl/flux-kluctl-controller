package v1alpha1

import (
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ObjectRef contains the information necessary to locate a resource within a cluster.
type ObjectRef struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func ConvertObjectRef(ref *k8s.ObjectRef) *ObjectRef {
	if ref == nil {
		return nil
	}
	return &ObjectRef{
		Group:     ref.GVK.Group,
		Version:   ref.GVK.Version,
		Kind:      ref.GVK.Kind,
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}

func ConvertResourceRefToKluctl(ref *ObjectRef) *k8s.ObjectRef {
	if ref == nil {
		return nil
	}
	return &k8s.ObjectRef{
		GVK: schema.GroupVersionKind{
			Group:   ref.Group,
			Version: ref.Version,
			Kind:    ref.Kind,
		},
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}

type FixedImage struct {
	Image         string     `json:"image"`
	ResultImage   string     `json:"resultImage"`
	DeployedImage *string    `json:"deployedImage,omitempty"`
	RegistryImage *string    `json:"registryImage,omitempty"`
	Namespace     *string    `json:"namespace,omitempty"`
	Object        *ObjectRef `json:"object,omitempty"`
	Deployment    *string    `json:"deployment,omitempty"`
	Container     *string    `json:"container,omitempty"`
	VersionFilter *string    `json:"versionFilter,omitempty"`
	DeployTags    []string   `json:"deployTags,omitempty"`
	DeploymentDir *string    `json:"deploymentDir,omitempty"`
}

func ConvertFixedImage(fi types.FixedImage) *FixedImage {
	return &FixedImage{
		Image:         fi.Image,
		ResultImage:   fi.ResultImage,
		DeployedImage: fi.DeployedImage,
		RegistryImage: fi.RegistryImage,
		Namespace:     fi.Namespace,
		Object:        ConvertObjectRef(fi.Object),
		Deployment:    fi.Deployment,
		Container:     fi.Container,
		VersionFilter: fi.VersionFilter,
		DeployTags:    fi.DeployTags,
		DeploymentDir: fi.DeploymentDir,
	}
}

func ConvertFixedImageToKluctl(fi FixedImage) types.FixedImage {
	return types.FixedImage{
		Image:         fi.Image,
		ResultImage:   fi.ResultImage,
		DeployedImage: fi.DeployedImage,
		RegistryImage: fi.RegistryImage,
		Namespace:     fi.Namespace,
		Object:        ConvertResourceRefToKluctl(fi.Object),
		Deployment:    fi.Deployment,
		Container:     fi.Container,
		VersionFilter: fi.VersionFilter,
		DeployTags:    fi.DeployTags,
		DeploymentDir: fi.DeploymentDir,
	}
}

func ConvertFixedImagesToKluctl(fi []FixedImage) []types.FixedImage {
	var ret []types.FixedImage
	for _, x := range fi {
		ret = append(ret, ConvertFixedImageToKluctl(x))
	}
	return ret
}

type Change struct {
	Type        string `json:"type"`
	JsonPath    string `json:"jsonPath"`
	UnifiedDiff string `json:"unifiedDiff,omitempty"`
}

type DeploymentError struct {
	Ref   ObjectRef `json:"ref"`
	Error string    `json:"error"`
}

type CommandResult struct {
	NewObjects     []ObjectRef       `json:"newObjects,omitempty"`
	ChangedObjects []ObjectRef       `json:"changedObjects,omitempty"`
	HookObjects    []ObjectRef       `json:"hookObjects,omitempty"`
	OrphanObjects  []ObjectRef       `json:"orphanObjects,omitempty"`
	DeletedObjects []ObjectRef       `json:"deletedObjects,omitempty"`
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
		ret.NewObjects = append(ret.NewObjects, *ConvertObjectRef(&x.Ref))
	}
	for _, x := range cmdResult.ChangedObjects {
		ret.ChangedObjects = append(ret.ChangedObjects, *ConvertObjectRef(&x.Ref))
	}
	for _, x := range cmdResult.HookObjects {
		ret.HookObjects = append(ret.HookObjects, *ConvertObjectRef(&x.Ref))
	}
	for _, x := range cmdResult.OrphanObjects {
		ret.OrphanObjects = append(ret.OrphanObjects, *ConvertObjectRef(&x))
	}
	for _, x := range cmdResult.DeletedObjects {
		ret.DeletedObjects = append(ret.DeletedObjects, *ConvertObjectRef(&x))
	}
	for _, x := range cmdResult.Errors {
		ret.Errors = append(ret.Errors, DeploymentError{Ref: *ConvertObjectRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.Warnings {
		ret.Warnings = append(ret.Warnings, DeploymentError{Ref: *ConvertObjectRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.SeenImages {
		ret.SeenImages = append(ret.SeenImages, *ConvertFixedImage(x))
	}
	return &ret
}

type ValidateResultEntry struct {
	Ref        ObjectRef `json:"ref"`
	Annotation string    `json:"annotation"`
	Message    string    `json:"message"`
}

type ValidateResult struct {
	Ready    bool                  `json:"ready"`
	Warnings []DeploymentError     `json:"warnings,omitempty"`
	Errors   []DeploymentError     `json:"errors,omitempty"`
	Results  []ValidateResultEntry `json:"results,omitempty"`
}

func ConvertValidateResult(cmdResult *types.ValidateResult) *ValidateResult {
	if cmdResult == nil {
		return nil
	}
	var ret ValidateResult
	ret.Ready = cmdResult.Ready
	for _, x := range cmdResult.Warnings {
		ret.Warnings = append(ret.Warnings, DeploymentError{Ref: *ConvertObjectRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.Errors {
		ret.Errors = append(ret.Errors, DeploymentError{Ref: *ConvertObjectRef(&x.Ref), Error: x.Error})
	}
	for _, x := range cmdResult.Results {
		ret.Results = append(ret.Results, ValidateResultEntry{Ref: *ConvertObjectRef(&x.Ref), Annotation: x.Annotation, Message: x.Message})
	}
	return &ret
}
