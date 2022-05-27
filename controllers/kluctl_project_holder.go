package controllers

import (
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/dependency"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KluctlProjectHolder interface {
	client.Object
	dependency.Dependent
	meta.ObjectWithConditions
	meta.ObjectWithConditionsSetter

	GetKluctlProject() *kluctlv1.KluctlProjectSpec
	GetKluctlTiming() *kluctlv1.KluctlTimingSpec
	GetKluctlStatus() *kluctlv1.KluctlProjectStatus
}

type KluctlProjectListHolder interface {
	client.ObjectList

	GetItems() []client.Object
}
