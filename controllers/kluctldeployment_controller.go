package controllers

import (
	"context"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
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
	source sourcev1.Source) (*ctrl.Result, error) {

	obj := objIn.(*kluctlv1.KluctlDeployment)

	pp, err := prepareProject(ctx, r.R, obj, source)
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}
	defer pp.cleanup()

	pt, err := pp.newTarget(obj.Spec.Target, obj.Spec.KluctlDeploymentTemplateSpec)
	if err != nil {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
		return nil, err
	}

	nextDeployTime := r.nextDeployTime(obj)
	nextValidateTime := r.nextValidateTime(obj)

	needDeploy := nextDeployTime.Before(time.Now()) || obj.Status.ObservedGeneration != obj.GetGeneration()
	needPrune := obj.Spec.Prune && needDeploy
	needValidate := needDeploy || nextValidateTime.Before(time.Now())

	obj.Status.ObservedGeneration = obj.GetGeneration()

	err = pt.withKluctlProjectTarget(ctx, func(targetContext *kluctl_project.TargetContext) error {
		obj.Status.CommonLabels = targetContext.DeploymentProject.GetCommonLabels()

		if needDeploy {
			// deploy the kluctl project
			deployResult, err := pt.kluctlDeploy(ctx, targetContext)
			kluctlv1.SetDeployResult(obj, pp.source.GetArtifact().Revision, deployResult, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		if needPrune {
			// run garbage collection for stale objects that do not have pruning disabled
			pruneResult, err := pt.kluctlPrune(ctx, targetContext)
			kluctlv1.SetPruneResult(obj, pp.source.GetArtifact().Revision, pruneResult, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PruneFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		if needValidate {
			validateResult, err := pt.kluctlValidate(ctx, targetContext)
			kluctlv1.SetValidateResult(obj, pp.source.GetArtifact().Revision, validateResult, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.ValidateFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		return nil
	})

	nextDeployTime = r.nextDeployTime(obj)
	nextValidateTime = r.nextValidateTime(obj)

	nextRunTime := nextDeployTime
	if nextValidateTime.Before(nextRunTime) {
		nextRunTime = nextValidateTime
	}

	ctrlResult := ctrl.Result{
		RequeueAfter: nextRunTime.Sub(time.Now()),
	}
	if ctrlResult.RequeueAfter < 0 {
		ctrlResult.RequeueAfter = 0
		ctrlResult.Requeue = true
	}

	deployOk := obj.Status.LastDeployResult != nil && obj.Status.LastDeployResult.Error == "" && len(obj.Status.LastDeployResult.Result.Errors) == 0
	pruneOk := obj.Status.LastPruneResult == nil || (obj.Status.LastPruneResult.Error == "" && len(obj.Status.LastPruneResult.Result.Errors) == 0)
	validateOk := obj.Status.LastValidateResult != nil && obj.Status.LastValidateResult.Error == "" && len(obj.Status.LastValidateResult.Result.Errors) == 0 && obj.Status.LastValidateResult.Result.Ready

	if deployOk && pruneOk && validateOk {
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionTrue, kluctlv1.ReconciliationSucceededReason, fmt.Sprintf("Deployed revision: %s", pp.source.GetArtifact().Revision), obj.GetGeneration(), pp.source.GetArtifact().Revision)
	}

	return &ctrlResult, nil
}

func (r *KluctlDeploymentReconcilerImpl) nextDeployTime(obj *kluctlv1.KluctlDeployment) time.Time {
	if obj.Status.LastDeployResult == nil {
		return time.Now()
	}

	nextRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.Spec.Interval.Duration)
	nextRetryRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.GetKluctlTiming().GetRetryInterval())
	if obj.Status.LastDeployResult != nil && obj.Status.LastDeployResult.Error != "" {
		nextRun = nextRetryRun
	}
	if obj.Status.LastDeployResult != nil && len(obj.Status.LastDeployResult.Result.Errors) != 0 {
		nextRun = nextRetryRun
	}

	return nextRun
}

func (r *KluctlDeploymentReconcilerImpl) nextValidateTime(obj *kluctlv1.KluctlDeployment) time.Time {
	if obj.Status.LastValidateResult == nil {
		return time.Now()
	}

	nextRun := obj.Status.LastValidateResult.AttemptedAt.Add(obj.Spec.ValidateInterval.Duration)
	return nextRun
}

func (r *KluctlDeploymentReconcilerImpl) Finalize(ctx context.Context, objIn KluctlProjectHolder) {
	log := ctrl.LoggerFrom(ctx)

	obj := objIn.(*kluctlv1.KluctlDeployment)

	if !obj.Spec.Prune || obj.Spec.Suspend {
		return
	}

	if len(obj.Status.CommonLabels) == 0 {
		log.V(1).Info("No commonLabels set, skipping deletion")
		return
	}

	log.V(1).Info("Deleting target")

	pp, err := prepareProject(ctx, r.R, obj, nil)
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
