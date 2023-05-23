package v1alpha1

import (
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
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
		Group:     ref.Group,
		Version:   ref.Version,
		Kind:      ref.Kind,
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}

func ConvertResourceRefToKluctl(ref *ObjectRef) *k8s.ObjectRef {
	if ref == nil {
		return nil
	}
	return &k8s.ObjectRef{
		Group:     ref.Group,
		Version:   ref.Version,
		Kind:      ref.Kind,
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
		Namespace:     fi.Namespace,
		Object:        ConvertObjectRef(fi.Object),
		Deployment:    fi.Deployment,
		Container:     fi.Container,
		DeployTags:    fi.DeployTags,
		DeploymentDir: fi.DeploymentDir,
	}
}

func ConvertFixedImageToKluctl(fi FixedImage) types.FixedImage {
	return types.FixedImage{
		Image:         fi.Image,
		ResultImage:   fi.ResultImage,
		DeployedImage: fi.DeployedImage,
		Namespace:     fi.Namespace,
		Object:        ConvertResourceRefToKluctl(fi.Object),
		Deployment:    fi.Deployment,
		Container:     fi.Container,
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
