package controllers

import (
	"context"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	types2 "github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type KluctlMultiDeploymentReconcilerImpl struct {
	R *KluctlProjectReconciler
}

// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctlmultideployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctlmultideployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctlmultideployments/finalizers,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets;gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets/status;gitrepositories/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *KluctlMultiDeploymentReconcilerImpl) NewObject() KluctlProjectHolder {
	return &kluctlv1.KluctlMultiDeployment{}
}

func (r *KluctlMultiDeploymentReconcilerImpl) NewObjectList() KluctlProjectListHolder {
	return &kluctlv1.KluctlMultiDeploymentList{}
}

func (r *KluctlMultiDeploymentReconcilerImpl) Reconcile(
	ctx context.Context,
	objIn KluctlProjectHolder,
	source sourcev1.Source) (*ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	obj := objIn.(*kluctlv1.KluctlMultiDeployment)

	pp, err := prepareProject(ctx, r.R, obj, source)
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}
	defer pp.cleanup()

	targets, err := pp.listTargets(ctx)
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}

	pattern, err := regexp.Compile(obj.Spec.TargetPattern)
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}

	kdList := &kluctlv1.KluctlDeploymentList{}
	err = r.R.List(ctx, kdList, client.InNamespace(obj.Namespace))
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}

	toDelete := make(map[string]*kluctlv1.KluctlDeployment)
	for _, x := range kdList.Items {
		x := x
		if !x.GetDeletionTimestamp().IsZero() {
			continue
		}
		for _, or := range x.OwnerReferences {
			if or.APIVersion == obj.APIVersion && or.Kind == obj.Kind && obj.Name == or.Name {
				toDelete[x.Spec.Target] = &x
				break
			}
		}
	}

	for _, target := range targets {
		if !pattern.MatchString(target.Name) {
			continue
		}

		delete(toDelete, target.Name)

		err := r.reconcileKluctlDeployment(ctx, obj, target)
		if err != nil {
			kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
			return nil, err
		}
	}

	for _, x := range toDelete {
		log.V(1).Info("Deleting KluctlDeployment %s", x.Name)

		err = r.R.Delete(ctx, x)
		if err != nil {
			kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
			return nil, err
		}
	}

	kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionTrue, kluctlv1.ReconciliationSucceededReason, fmt.Sprintf("Reconciled revision: %s", pp.source.GetArtifact().Revision), obj.GetGeneration(), pp.source.GetArtifact().Revision)
	return nil, nil
}

func (r *KluctlMultiDeploymentReconcilerImpl) reconcileKluctlDeployment(ctx context.Context, obj *kluctlv1.KluctlMultiDeployment, target *types2.Target) error {
	log := ctrl.LoggerFrom(ctx)

	baseName := fmt.Sprintf("%s-%s", obj.Name, target.Name)
	kd := &kluctlv1.KluctlDeployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", baseName, utils.Sha256String(baseName)[:8]),
			Namespace: obj.Namespace,
		},
	}

	mres, err := controllerutil.CreateOrUpdate(ctx, r.R.Client, kd, func() error {
		if err := controllerutil.SetControllerReference(obj, kd, r.R.Scheme); err != nil {
			return err
		}

		kd.Spec = kluctlv1.KluctlDeploymentSpec{
			KluctlProjectSpec:            obj.Spec.KluctlProjectSpec,
			KluctlDeploymentTemplateSpec: obj.Spec.Template.Spec,
			Target:                       target.Name,
		}
		return nil
	})
	if err != nil {
		return err
	}

	if mres != controllerutil.OperationResultNone {
		log.V(1).Info("CreateOrUpdate returned %s", mres)
	}

	return nil
}

func (r *KluctlMultiDeploymentReconcilerImpl) Finalize(ctx context.Context, obj KluctlProjectHolder) {
}
