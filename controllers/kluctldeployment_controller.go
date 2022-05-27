package controllers

import (
	"context"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KluctlDeploymentReconcilerImpl struct {
	R *KluctlProjectReconciler
}

// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments/finalizers,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets;gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets/status;gitrepositories/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *KluctlDeploymentReconcilerImpl) NewObject() KluctlProjectHolder {
	return &kluctlv1.KluctlDeployment{}
}

func (r *KluctlDeploymentReconcilerImpl) NewObjectList() KluctlProjectListHolder {
	return &kluctlv1.KluctlDeploymentList{}
}

func (r *KluctlDeploymentReconcilerImpl) Reconcile(
	ctx context.Context,
	objIn KluctlProjectHolder,
	source sourcev1.Source) error {

	obj := objIn.(*kluctlv1.KluctlDeployment)

	pp, err := prepareProject(ctx, r.R, obj, source)
	if err != nil {
		kluctlv1.SetKluctlDeploymentReadiness(obj, metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), pp.source.GetArtifact().Revision, nil, nil)
		return err
	}
	defer pp.cleanup()

	pt, err := pp.newTarget(obj.Spec.Target, obj.Spec.KluctlDeploymentTemplateSpec)
	if err != nil {
		kluctlv1.SetKluctlDeploymentReadiness(obj, metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), pp.source.GetArtifact().Revision, nil, nil)
		return err
	}

	var deployResult *kluctlv1.CommandResult
	var pruneResult *kluctlv1.CommandResult

	err = pt.withKluctlProjectTarget(ctx, func(targetContext *kluctl_project.TargetContext) error {
		obj.Status.CommonLabels = targetContext.DeploymentProject.GetCommonLabels()

		// deploy the kluctl project
		deployResult, err = pt.kluctlDeploy(ctx, targetContext)
		if err != nil {
			kluctlv1.SetKluctlDeploymentReadiness(obj, metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), pp.source.GetArtifact().Revision, deployResult, nil)
			return err
		}

		if obj.Spec.Prune {
			// run garbage collection for stale objects that do not have pruning disabled
			pruneResult, err = pt.kluctlPrune(ctx, targetContext)
			if err != nil {
				kluctlv1.SetKluctlDeploymentReadiness(obj, metav1.ConditionFalse, kluctlv1.PruneFailedReason, err.Error(), pp.source.GetArtifact().Revision, deployResult, pruneResult)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	kluctlv1.SetKluctlDeploymentReadiness(obj, metav1.ConditionTrue, kluctlv1.ReconciliationSucceededReason, fmt.Sprintf("Deployed revision: %s", pp.source.GetArtifact().Revision), pp.source.GetArtifact().Revision, deployResult, pruneResult)
	return nil
}

func (r *KluctlDeploymentReconcilerImpl) Finalize(ctx context.Context, objIn KluctlProjectHolder) {
	log := ctrl.LoggerFrom(ctx)

	obj := objIn.(*kluctlv1.KluctlDeployment)

	if !obj.Spec.Prune || obj.Spec.Suspend {
		return
	}

	if len(obj.Status.CommonLabels) != 0 {
		log.V(1).Info("No commonLabels set, skipping deletion")
		return
	}

	log.V(1).Info("Deleting target")

	source, err := r.R.getSource(ctx, obj)
	if err != nil {
		return
	}
	pp, err := prepareProject(ctx, r.R, obj, source)
	if err != nil {
		return
	}
	defer pp.cleanup()

	pt, err := pp.newTarget(obj.Spec.Target, obj.Spec.KluctlDeploymentTemplateSpec)
	if err != nil {
		return
	}

	_, _ = pt.kluctlDelete(ctx, obj.Status.CommonLabels)
}
