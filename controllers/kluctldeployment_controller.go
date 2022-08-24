package controllers

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	project "github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/utils/uo"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sort"
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

	err = pt.withKluctlProjectTarget(ctx, func(targetContext *kluctl_project.TargetContext) error {
		obj.Status.CommonLabels = targetContext.DeploymentProject.GetCommonLabels()
		obj.Status.SetRawTarget(targetContext.Target)

		objectsHash := r.calcObjectsHash(targetContext)

		needDeploy := (nextDeployTime != nil && nextDeployTime.Before(time.Now())) || obj.Status.ObservedGeneration != obj.GetGeneration()
		if obj.Status.LastDeployResult == nil || obj.Status.LastDeployResult.ObjectsHash != objectsHash {
			needDeploy = true
		}
		needPrune := obj.Spec.Prune && needDeploy
		needValidate := needDeploy || nextValidateTime.Before(time.Now())

		if needDeploy {
			// deploy the kluctl project
			var deployResult *types.CommandResult
			if obj.Spec.DeployMode == kluctlv1.KluctlDeployModeFull {
				deployResult, err = pt.kluctlDeploy(ctx, targetContext)
			} else if obj.Spec.DeployMode == kluctlv1.KluctlDeployPokeImages {
				deployResult, err = pt.kluctlPokeImages(ctx, targetContext)
			} else {
				err = fmt.Errorf("deployMode '%s' not supported", obj.Spec.DeployMode)
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
			kluctlv1.SetDeployResult(obj, pp.source.GetArtifact().Revision, deployResult, objectsHash, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		if needPrune {
			// run garbage collection for stale objects that do not have pruning disabled
			pruneResult, err := pt.kluctlPrune(ctx, targetContext)
			kluctlv1.SetPruneResult(obj, pp.source.GetArtifact().Revision, pruneResult, objectsHash, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.PruneFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		if needValidate {
			validateResult, err := pt.kluctlValidate(ctx, targetContext)
			kluctlv1.SetValidateResult(obj, pp.source.GetArtifact().Revision, validateResult, objectsHash, err)
			if err != nil {
				kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.ValidateFailedReason, err.Error(), obj.GetGeneration(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		return nil
	})

	nextDeployTime = r.nextDeployTime(obj)
	nextValidateTime = r.nextValidateTime(obj)

	nextRunTime := time.Now().Add(obj.Spec.Interval.Duration)
	if nextDeployTime != nil && nextDeployTime.Before(nextRunTime) {
		nextRunTime = *nextDeployTime
	}
	if nextValidateTime != nil && nextValidateTime.Before(nextRunTime) {
		nextRunTime = *nextValidateTime
	}

	ctrlResult := ctrl.Result{
		RequeueAfter: nextRunTime.Sub(time.Now()),
	}
	if ctrlResult.RequeueAfter < 0 {
		ctrlResult.RequeueAfter = 0
		ctrlResult.Requeue = true
	}

	lastDeployResult := obj.Status.LastDeployResult.ParseResult()
	lastPruneResult := obj.Status.LastPruneResult.ParseResult()
	lastValidateResult := obj.Status.LastValidateResult.ParseResult()

	deployOk := lastDeployResult != nil && obj.Status.LastDeployResult.Error == "" && len(lastDeployResult.Errors) == 0
	pruneOk := lastPruneResult == nil || (obj.Status.LastPruneResult.Error == "" && len(lastPruneResult.Errors) == 0)
	validateOk := lastValidateResult != nil && obj.Status.LastValidateResult.Error == "" && len(lastValidateResult.Errors) == 0 && lastValidateResult.Ready

	finalStatus := ""

	if obj.Status.LastDeployResult != nil {
		finalStatus += "deploy: "
		if deployOk {
			finalStatus += "ok"
		} else {
			finalStatus += "failed"
		}
	}
	if obj.Status.LastPruneResult != nil {
		if finalStatus != "" {
			finalStatus += ", "
		}
		finalStatus += "prune: "
		if pruneOk {
			finalStatus += "ok"
		} else {
			finalStatus += "failed"
		}
	}
	if obj.Status.LastValidateResult != nil {
		if finalStatus != "" {
			finalStatus += ", "
		}
		finalStatus += "validate: "
		if validateOk {
			finalStatus += "ok"
		} else {
			finalStatus += "failed"
		}
	}

	var conditionStatus metav1.ConditionStatus
	var reason string
	if deployOk && pruneOk {
		if validateOk {
			conditionStatus = metav1.ConditionTrue
			reason = kluctlv1.ReconciliationSucceededReason
		} else {
			conditionStatus = metav1.ConditionFalse
			reason = kluctlv1.ValidateFailedReason
		}
	} else {
		conditionStatus = metav1.ConditionFalse
		reason = kluctlv1.DeployFailedReason
	}
	kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), conditionStatus, reason, finalStatus, obj.GetGeneration(), pp.source.GetArtifact().Revision)

	obj.Status.ObservedGeneration = obj.GetGeneration()

	return &ctrlResult, err
}

func (r *KluctlDeploymentReconcilerImpl) nextDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	t1 := r.nextRequestedDeployTime(obj)
	t2 := r.nextRetryDeployTime(obj)
	if t1 != nil && t2 == nil {
		return t1
	} else if t1 == nil && t2 != nil {
		return t2
	} else if t1 != nil && t2 != nil {
		if t1.Before(*t2) {
			return t1
		} else {
			return t2
		}
	}
	return nil
}

func (r *KluctlDeploymentReconcilerImpl) nextRequestedDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	v, ok := obj.Annotations[kluctlv1.KluctlDeployRequestAnnotation]
	if !ok {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	if obj.Status.LastDeployResult == nil || obj.Status.LastDeployResult.AttemptedAt.Time.Before(t) {
		return &t
	}
	return nil
}

func (r *KluctlDeploymentReconcilerImpl) nextRetryDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	lastDeployResult := obj.Status.LastDeployResult.ParseResult()

	if lastDeployResult == nil {
		now := time.Now()
		return &now
	}

	if obj.Status.LastDeployResult.Error != "" || len(lastDeployResult.Errors) != 0 {
		nextRetryRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.GetKluctlTiming().GetRetryInterval())
		return &nextRetryRun
	}

	if obj.Spec.DeployInterval == nil {
		return nil
	}

	nextRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.Spec.DeployInterval.Duration)
	return &nextRun
}

func (r *KluctlDeploymentReconcilerImpl) nextValidateTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	if obj.Status.LastValidateResult == nil {
		now := time.Now()
		return &now
	}

	nextRun := obj.Status.LastValidateResult.AttemptedAt.Add(obj.Spec.ValidateInterval.Duration)
	return &nextRun
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

func (r *KluctlDeploymentReconcilerImpl) calcObjectsHash(targetContext *project.TargetContext) string {
	h := sha256.New()
	tw := tar.NewWriter(h)
	var objects []any
	for _, di := range targetContext.DeploymentCollection.Deployments {
		for _, o := range di.Objects {
			objects = append(objects, o)
		}
	}
	sort.Slice(objects, func(i, j int) bool {
		a := objects[i].(*uo.UnstructuredObject)
		b := objects[i].(*uo.UnstructuredObject)
		return a.GetK8sRef().String() < b.GetK8sRef().String()
	})
	err := yaml.WriteYamlAllStream(h, objects)
	if err != nil {
		panic(err)
	}
	err = tw.Flush()
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
