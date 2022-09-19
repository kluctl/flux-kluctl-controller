package controllers

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	apiacl "github.com/fluxcd/pkg/apis/acl"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/acl"
	"github.com/fluxcd/pkg/runtime/events"
	"github.com/fluxcd/pkg/runtime/metrics"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/hashicorp/go-retryablehttp"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	ssh_pool "github.com/kluctl/kluctl/v2/pkg/git/ssh-pool"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	project "github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	"github.com/kluctl/kluctl/v2/pkg/status"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/utils/uo"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kuberecorder "k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
	"time"
)

type KluctlDeploymentReconciler struct {
	client.Client
	httpClient            *retryablehttp.Client
	requeueDependency     time.Duration
	Scheme                *runtime.Scheme
	EventRecorder         kuberecorder.EventRecorder
	MetricsRecorder       *metrics.Recorder
	ControllerName        string
	statusManager         string
	NoCrossNamespaceRefs  bool
	DefaultServiceAccount string

	SshPool *ssh_pool.SshPool
}

// KluctlDeploymentReconcilerOpts contains options for the BaseReconciler.
type KluctlDeploymentReconcilerOpts struct {
	MaxConcurrentReconciles int
	HTTPRetry               int
}

// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flux.kluctl.io,resources=kluctldeployments/finalizers,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets;gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets/status;gitrepositories/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *KluctlDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reconcileStart := time.Now()

	ctx = status.NewContext(ctx, status.NewSimpleStatusHandler(func(message string) {
		log.Info(message)
	}, false, false))

	obj := &kluctlv1.KluctlDeployment{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, obj.Spec.GetTimeout())
	defer cancel()

	retryInterval := obj.Spec.GetRetryInterval()
	interval := obj.Spec.Interval.Duration

	// Record suspended status metric
	defer r.recordSuspension(ctx, obj)

	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(obj, kluctlv1.KluctlDeploymentFinalizer) {
		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.AddFinalizer(obj, kluctlv1.KluctlDeploymentFinalizer)
		if err := r.Patch(ctx, obj, patch, client.FieldOwner(r.statusManager)); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}

	// Examine if the object is under deletion
	if !obj.GetDeletionTimestamp().IsZero() {
		return r.finalize(ctx, obj)
	}

	// Return early if the KluctlDeployment is suspended.
	if obj.Spec.Suspend {
		log.Info("Reconciliation is suspended for this object")
		return ctrl.Result{}, nil
	}

	// resolve source reference
	source, err := r.getSource(ctx, obj.Spec.SourceRef, obj.GetNamespace(), r.NoCrossNamespaceRefs)
	if err != nil {
		if apierrors.IsNotFound(err) {
			msg := fmt.Sprintf("Source '%s' not found", obj.Spec.SourceRef)
			patch := client.MergeFrom(obj.DeepCopy())
			setReadiness(obj, metav1.ConditionFalse, kluctlv1.ArtifactFailedReason, msg)
			if err := r.Status().Patch(ctx, obj, patch); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			r.recordReadiness(ctx, obj)
			log.Info(msg)
			// do not requeue immediately, when the source is created the watcher should trigger a reconciliation
			return ctrl.Result{RequeueAfter: retryInterval}, nil
		}

		if acl.IsAccessDenied(err) {
			patch := client.MergeFrom(obj.DeepCopy())
			setReadiness(obj, metav1.ConditionFalse, apiacl.AccessDeniedReason, err.Error())
			if err := r.Status().Patch(ctx, obj, patch); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			log.Error(err, "access denied to cross-namespace source")
			r.recordReadiness(ctx, obj)
			r.event(ctx, obj, "unknown", events.EventSeverityError, err.Error(), nil)
			return ctrl.Result{RequeueAfter: retryInterval}, nil
		}

		// retry on transient errors
		return ctrl.Result{Requeue: true}, err
	}

	if source.GetArtifact() == nil {
		msg := "Source is not ready, artifact not found"
		patch := client.MergeFrom(obj.DeepCopy())
		setReadiness(obj, metav1.ConditionFalse, kluctlv1.ArtifactFailedReason, msg)
		if err := r.Status().Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		r.recordReadiness(ctx, obj)
		log.Info(msg)
		// do not requeue immediately, when the artifact is created the watcher should trigger a reconciliation
		return ctrl.Result{RequeueAfter: retryInterval}, nil
	}

	sourceRevision := source.GetArtifact().Revision

	// record reconciliation duration
	if r.MetricsRecorder != nil {
		objRef, err := reference.GetReference(r.Scheme, obj)
		if err != nil {
			return ctrl.Result{}, err
		}
		defer r.MetricsRecorder.RecordDuration(*objRef, reconcileStart)
	}

	// set the reconciliation status to progressing
	if obj.Status.ObservedGeneration == 0 {
		patch := client.MergeFrom(obj.DeepCopy())
		setReadiness(obj, metav1.ConditionUnknown, meta.ProgressingReason, "reconciliation in progress")
		if err := r.Status().Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		r.recordReadiness(ctx, obj)
	}

	// record the value of the reconciliation request, if any
	if v, ok := meta.ReconcileAnnotationValue(obj.GetAnnotations()); ok {
		obj.Status.SetLastHandledReconcileRequest(v)
	}

	// reconcile kluctlDeployment by applying the latest revision
	patch := client.MergeFrom(obj.DeepCopy())
	ctrlResult, reconcileErr := r.doReconcile(ctx, obj, source)
	if err := r.Status().Patch(ctx, obj, patch); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	r.recordReadiness(ctx, obj)

	if ctrlResult == nil {
		if reconcileErr != nil {
			ctrlResult = &ctrl.Result{RequeueAfter: retryInterval}
		} else {
			ctrlResult = &ctrl.Result{RequeueAfter: interval}
		}
	}

	// broadcast the reconciliation failure and requeue at the specified retry interval
	if reconcileErr != nil {
		log.Error(reconcileErr, fmt.Sprintf("Reconciliation failed after %s, next try in %s",
			time.Since(reconcileStart).String(),
			ctrlResult.RequeueAfter.String()),
			"revision",
			sourceRevision)
		r.event(ctx, obj, sourceRevision, events.EventSeverityError,
			reconcileErr.Error(), nil)
		return *ctrlResult, nil
	}

	// broadcast the reconciliation result and requeue at the specified interval
	msg := fmt.Sprintf("Reconciliation finished in %s, next run in %s",
		time.Since(reconcileStart).String(),
		ctrlResult.RequeueAfter.String())
	log.Info(msg, "revision", sourceRevision)
	r.event(ctx, obj, sourceRevision, events.EventSeverityInfo,
		msg, map[string]string{kluctlv1.GroupVersion.Group + "/commit_status": "update"})
	return *ctrlResult, nil
}

func (r *KluctlDeploymentReconciler) doReconcile(
	ctx context.Context,
	obj *kluctlv1.KluctlDeployment,
	source sourcev1.Source) (*ctrl.Result, error) {

	mustDownscale, err := r.calcDownTime(obj)
	if err != nil {
		return nil, err
	}
	isDownscaled := obj.Status.DownscaledAt != nil

	pp, err := prepareProject(ctx, r, obj, source)
	if err != nil {
		setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), pp.source.GetArtifact().Revision)
		return nil, err
	}
	defer pp.cleanup()

	pt, err := pp.newTarget(obj.Spec.Target, obj.Spec.KluctlDeploymentTemplateSpec)
	if err != nil {
		setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.PrepareFailedReason, err.Error(), pp.source.GetArtifact().Revision)
		return nil, err
	}

	err = pt.withKluctlProjectTarget(ctx, func(targetContext *kluctl_project.TargetContext) error {
		obj.Status.CommonLabels = targetContext.DeploymentProject.GetCommonLabels()
		obj.Status.SetRawTarget(targetContext.Target)

		objectsHash := r.calcObjectsHash(targetContext)

		needDeploy := false
		needPrune := false
		needDownscale := false
		needValidate := false
		if mustDownscale {
			needDownscale = !isDownscaled
		} else {
			nextDeployTime := r.nextDeployTime(obj)
			nextValidateTime := r.nextValidateTime(obj)
			needDeploy = (nextDeployTime != nil && nextDeployTime.Before(time.Now())) || obj.Status.ObservedGeneration != obj.GetGeneration()
			if obj.Status.LastDeployResult == nil || obj.Status.LastDeployResult.ObjectsHash != objectsHash {
				needDeploy = true
			}

			needPrune = obj.Spec.Prune && needDeploy
			needValidate = !mustDownscale && (needDeploy || nextValidateTime.Before(time.Now()))
		}

		if needDeploy {
			// deploy the kluctl project
			var deployResult *types.CommandResult
			if obj.Spec.DeployMode == kluctlv1.KluctlDeployModeFull {
				deployResult, err = pt.kluctlDeploy(ctx, targetContext)
			} else if obj.Spec.DeployMode == kluctlv1.KluctlDeployPokeImages {
				deployResult, err = pt.kluctlPokeImages(ctx, targetContext)
			} else {
				err = fmt.Errorf("deployMode '%s' not supported", obj.Spec.DeployMode)
				setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), pp.source.GetArtifact().Revision)
				return err
			}
			kluctlv1.SetDeployResult(obj, pp.source.GetArtifact().Revision, deployResult, objectsHash, err)
			if err != nil {
				setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.DeployFailedReason, err.Error(), pp.source.GetArtifact().Revision)
				return err
			}
			obj.Status.DownscaledAt = nil
		}

		if needPrune {
			// run garbage collection for stale objects that do not have pruning disabled
			pruneResult, err := pt.kluctlPrune(ctx, targetContext)
			kluctlv1.SetPruneResult(obj, pp.source.GetArtifact().Revision, pruneResult, objectsHash, err)
			if err != nil {
				setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.PruneFailedReason, err.Error(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		if needDownscale {
			downscaleResult, err := pt.kluctlDownscale(ctx, targetContext)
			kluctlv1.SetDownscaleResult(obj, pp.source.GetArtifact().Revision, downscaleResult, objectsHash, err)
			if err != nil {
				setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.DownscaleFailedReason, err.Error(), pp.source.GetArtifact().Revision)
				return err
			}
			t := metav1.Now()
			obj.Status.DownscaledAt = &t
		}

		if needValidate {
			validateResult, err := pt.kluctlValidate(ctx, targetContext)
			kluctlv1.SetValidateResult(obj, pp.source.GetArtifact().Revision, validateResult, objectsHash, err)
			if err != nil {
				setReadinessWithRevision(obj, metav1.ConditionFalse, kluctlv1.ValidateFailedReason, err.Error(), pp.source.GetArtifact().Revision)
				return err
			}
		}

		return nil
	})
	obj.Status.ObservedGeneration = obj.GetGeneration()
	if v, ok := obj.GetAnnotations()[kluctlv1.KluctlDeployRequestAnnotation]; ok {
		obj.Status.LastHandledDeployAt = v
	}
	if err != nil {
		return nil, err
	}

	finalStatus, reason := r.buildFinalStatus(obj)
	if reason != kluctlv1.ReconciliationSucceededReason {
		setReadinessWithRevision(obj, metav1.ConditionFalse, reason, finalStatus, pp.source.GetArtifact().Revision)
		return nil, fmt.Errorf(finalStatus)
	}
	setReadinessWithRevision(obj, metav1.ConditionTrue, reason, finalStatus, pp.source.GetArtifact().Revision)

	var ctrlResult ctrl.Result
	ctrlResult.RequeueAfter = r.nextReconcileTime(obj).Sub(time.Now())
	if ctrlResult.RequeueAfter < 0 {
		ctrlResult.RequeueAfter = 0
		ctrlResult.Requeue = true
	}
	return &ctrlResult, nil
}

func (r *KluctlDeploymentReconciler) calcDownTime(obj *kluctlv1.KluctlDeployment) (bool, error) {
	var isUpTime bool
	if obj.Spec.Downscale != nil {
		isUpTime = false
		if obj.Spec.Downscale.Enabled && len(obj.Spec.Downscale.UpTime) == 0 && len(obj.Spec.Downscale.DownTime) == 0 {
			return false, fmt.Errorf("either upTime or downTime must be specified when downscaling is enabled")
		}
		for _, s := range obj.Spec.Downscale.UpTime {
			x, err := MatchesTimeSpec(time.Now(), s)
			if err != nil {
				return false, err
			}
			if x {
				isUpTime = true
			}
		}
		if len(obj.Spec.Downscale.UpTime) == 0 {
			isUpTime = true
		}
		for _, s := range obj.Spec.Downscale.DownTime {
			x, err := MatchesTimeSpec(time.Now(), s)
			if err != nil {
				return false, err
			}
			if x {
				isUpTime = false
			}
		}
	} else {
		isUpTime = true
	}
	return !isUpTime, nil
}

func (r *KluctlDeploymentReconciler) buildFinalStatus(obj *kluctlv1.KluctlDeployment) (finalStatus string, reason string) {
	lastDeployResult := obj.Status.LastDeployResult.ParseResult()
	lastPruneResult := obj.Status.LastPruneResult.ParseResult()
	lastValidateResult := obj.Status.LastValidateResult.ParseResult()

	deployOk := lastDeployResult != nil && obj.Status.LastDeployResult.Error == "" && len(lastDeployResult.Errors) == 0
	pruneOk := lastPruneResult == nil || (obj.Status.LastPruneResult.Error == "" && len(lastPruneResult.Errors) == 0)
	validateOk := lastValidateResult != nil && obj.Status.LastValidateResult.Error == "" && len(lastValidateResult.Errors) == 0 && lastValidateResult.Ready

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

	if deployOk && pruneOk {
		if validateOk {
			reason = kluctlv1.ReconciliationSucceededReason
		} else {
			reason = kluctlv1.ValidateFailedReason
			return
		}
	} else {
		reason = kluctlv1.DeployFailedReason
		return
	}
	return
}

func (r *KluctlDeploymentReconciler) nextReconcileTime(obj *kluctlv1.KluctlDeployment) time.Time {
	nextDeployTime := r.nextDeployTime(obj)
	nextValidateTime := r.nextValidateTime(obj)

	nextRunTime := time.Now().Add(obj.Spec.Interval.Duration)
	if nextDeployTime != nil && nextDeployTime.Before(nextRunTime) {
		nextRunTime = *nextDeployTime
	}
	if nextValidateTime != nil && nextValidateTime.Before(nextRunTime) {
		nextRunTime = *nextValidateTime
	}

	return nextRunTime
}

func (r *KluctlDeploymentReconciler) nextDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
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

func (r *KluctlDeploymentReconciler) nextRequestedDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	v, ok := obj.Annotations[kluctlv1.KluctlDeployRequestAnnotation]
	if !ok {
		return nil
	}
	if v != obj.Status.LastHandledDeployAt {
		t := time.Now()
		return &t
	}
	return nil
}

func (r *KluctlDeploymentReconciler) nextRetryDeployTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	lastDeployResult := obj.Status.LastDeployResult.ParseResult()

	if lastDeployResult == nil {
		now := time.Now()
		return &now
	}

	if obj.Status.LastDeployResult.Error != "" || len(lastDeployResult.Errors) != 0 {
		nextRetryRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.Spec.GetRetryInterval())
		return &nextRetryRun
	}

	if obj.Spec.DeployInterval == nil {
		return nil
	}

	nextRun := obj.Status.LastDeployResult.AttemptedAt.Add(obj.Spec.DeployInterval.Duration)
	return &nextRun
}

func (r *KluctlDeploymentReconciler) nextValidateTime(obj *kluctlv1.KluctlDeployment) *time.Time {
	if obj.Status.LastValidateResult == nil {
		now := time.Now()
		return &now
	}

	nextRun := obj.Status.LastValidateResult.AttemptedAt.Add(obj.Spec.ValidateInterval.Duration)
	return &nextRun
}

func (r *KluctlDeploymentReconciler) finalize(ctx context.Context, obj *kluctlv1.KluctlDeployment) (ctrl.Result, error) {
	r.doFinalize(ctx, obj)

	// Record deleted status
	r.recordReadiness(ctx, obj)

	// Remove our finalizer from the list and update it
	controllerutil.RemoveFinalizer(obj, kluctlv1.KluctlDeploymentFinalizer)
	if err := r.Update(ctx, obj, client.FieldOwner(r.statusManager)); err != nil {
		return ctrl.Result{}, err
	}

	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}

func (r *KluctlDeploymentReconciler) doFinalize(ctx context.Context, obj *kluctlv1.KluctlDeployment) {
	log := ctrl.LoggerFrom(ctx)

	if !obj.Spec.Prune || obj.Spec.Suspend {
		return
	}

	if len(obj.Status.CommonLabels) == 0 {
		log.V(1).Info("No commonLabels set, skipping deletion")
		return
	}

	log.V(1).Info("Deleting target")

	pp, err := prepareProject(ctx, r, obj, nil)
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

func (r *KluctlDeploymentReconciler) calcObjectsHash(targetContext *project.TargetContext) string {
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
