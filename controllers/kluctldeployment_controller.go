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

package controllers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kuberecorder "k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apiacl "github.com/fluxcd/pkg/apis/acl"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/acl"
	"github.com/fluxcd/pkg/runtime/events"
	"github.com/fluxcd/pkg/runtime/metrics"
	"github.com/fluxcd/pkg/runtime/predicates"
	"github.com/fluxcd/pkg/untar"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"

	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"

	_ "github.com/kluctl/kluctl/v2/pkg/jinja2"
)

// KluctlDeploymentReconciler reconciles a KluctlDeployment object
type KluctlDeploymentReconciler struct {
	client.Client
	httpClient           *retryablehttp.Client
	requeueDependency    time.Duration
	Scheme               *runtime.Scheme
	EventRecorder        kuberecorder.EventRecorder
	MetricsRecorder      *metrics.Recorder
	ControllerName       string
	statusManager        string
	NoCrossNamespaceRefs bool
}

// +kubebuilder:rbac:groups=kluctl.io,resources=kluctldeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kluctl.io,resources=kluctldeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kluctl.io,resources=kluctldeployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets;gitrepositories,verbs=get;list;watch
// +kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=buckets/status;gitrepositories/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// KluctlDeploymentReconcilerOptions contains options for the KluctlDeploymentReconciler.
type KluctlDeploymentReconcilerOptions struct {
	MaxConcurrentReconciles   int
	HTTPRetry                 int
	DependencyRequeueInterval time.Duration
}

// SetupWithManager sets up the controller with the Manager.
func (r *KluctlDeploymentReconciler) SetupWithManager(mgr ctrl.Manager, opts KluctlDeploymentReconcilerOptions) error {
	const (
		gitRepositoryIndexKey string = ".metadata.gitRepository"
		bucketIndexKey        string = ".metadata.bucket"
	)

	// Index the KluctlDeployments by the GitRepository references they (may) point at.
	if err := mgr.GetCache().IndexField(context.TODO(), &kluctlv1.KluctlDeployment{}, gitRepositoryIndexKey,
		r.indexBy(sourcev1.GitRepositoryKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	// Index the KluctlDeployments by the Bucket references they (may) point at.
	if err := mgr.GetCache().IndexField(context.TODO(), &kluctlv1.KluctlDeployment{}, bucketIndexKey,
		r.indexBy(sourcev1.BucketKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	r.requeueDependency = opts.DependencyRequeueInterval
	r.statusManager = fmt.Sprintf("gotk-%s", r.ControllerName)

	// Configure the retryable http client used for fetching artifacts.
	// By default it retries 10 times within a 3.5 minutes window.
	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = 5 * time.Second
	httpClient.RetryWaitMax = 30 * time.Second
	httpClient.RetryMax = opts.HTTPRetry
	httpClient.Logger = nil
	r.httpClient = httpClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&kluctlv1.KluctlDeployment{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicates.ReconcileRequestedPredicate{}),
		)).
		Watches(
			&source.Kind{Type: &sourcev1.GitRepository{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForRevisionChangeOf(gitRepositoryIndexKey)),
			builder.WithPredicates(SourceRevisionChangePredicate{}),
		).
		Watches(
			&source.Kind{Type: &sourcev1.Bucket{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForRevisionChangeOf(bucketIndexKey)),
			builder.WithPredicates(SourceRevisionChangePredicate{}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}

func (r *KluctlDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reconcileStart := time.Now()

	var kluctlDeployment kluctlv1.KluctlDeployment
	if err := r.Get(ctx, req.NamespacedName, &kluctlDeployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Record suspended status metric
	defer r.recordSuspension(ctx, kluctlDeployment)

	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&kluctlDeployment, kluctlv1.KluctlDeploymentFinalizer) {
		patch := client.MergeFrom(kluctlDeployment.DeepCopy())
		controllerutil.AddFinalizer(&kluctlDeployment, kluctlv1.KluctlDeploymentFinalizer)
		if err := r.Patch(ctx, &kluctlDeployment, patch, client.FieldOwner(r.statusManager)); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}

	// Examine if the object is under deletion
	if !kluctlDeployment.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.finalize(ctx, kluctlDeployment)
	}

	// Return early if the KluctlDeployment is suspended.
	if kluctlDeployment.Spec.Suspend {
		log.Info("Reconciliation is suspended for this object")
		return ctrl.Result{}, nil
	}

	// resolve source reference
	source, err := r.getSource(ctx, kluctlDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			msg := fmt.Sprintf("Source '%s' not found", kluctlDeployment.Spec.SourceRef.String())
			kluctlDeployment = kluctlv1.KluctlDeploymentNotReady(kluctlDeployment, "", "", kluctlv1.ArtifactFailedReason, msg)
			if err := r.patchStatus(ctx, req, kluctlDeployment.Status); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			r.recordReadiness(ctx, kluctlDeployment)
			log.Info(msg)
			// do not requeue immediately, when the source is created the watcher should trigger a reconciliation
			return ctrl.Result{RequeueAfter: kluctlDeployment.GetRetryInterval()}, nil
		}

		if acl.IsAccessDenied(err) {
			kluctlDeployment = kluctlv1.KluctlDeploymentNotReady(kluctlDeployment, "", "", apiacl.AccessDeniedReason, err.Error())
			if err := r.patchStatus(ctx, req, kluctlDeployment.Status); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			log.Error(err, "access denied to cross-namespace source")
			r.recordReadiness(ctx, kluctlDeployment)
			r.event(ctx, kluctlDeployment, "unknown", events.EventSeverityError, err.Error(), nil)
			return ctrl.Result{RequeueAfter: kluctlDeployment.GetRetryInterval()}, nil
		}

		// retry on transient errors
		return ctrl.Result{Requeue: true}, err
	}

	if source.GetArtifact() == nil {
		msg := "Source is not ready, artifact not found"
		kluctlDeployment = kluctlv1.KluctlDeploymentNotReady(kluctlDeployment, "", "", kluctlv1.ArtifactFailedReason, msg)
		if err := r.patchStatus(ctx, req, kluctlDeployment.Status); err != nil {
			log.Error(err, "unable to update status for artifact not found")
			return ctrl.Result{Requeue: true}, err
		}
		r.recordReadiness(ctx, kluctlDeployment)
		log.Info(msg)
		// do not requeue immediately, when the artifact is created the watcher should trigger a reconciliation
		return ctrl.Result{RequeueAfter: kluctlDeployment.GetRetryInterval()}, nil
	}

	// check dependencies
	if len(kluctlDeployment.Spec.DependsOn) > 0 {
		if err := r.checkDependencies(source, kluctlDeployment); err != nil {
			kluctlDeployment = kluctlv1.KluctlDeploymentNotReady(
				kluctlDeployment, source.GetArtifact().Revision, "", kluctlv1.DependencyNotReadyReason, err.Error())
			if err := r.patchStatus(ctx, req, kluctlDeployment.Status); err != nil {
				log.Error(err, "unable to update status for dependency not ready")
				return ctrl.Result{Requeue: true}, err
			}
			// we can't rely on exponential backoff because it will prolong the execution too much,
			// instead we requeue on a fix interval.
			msg := fmt.Sprintf("Dependencies do not meet ready condition, retrying in %s", r.requeueDependency.String())
			log.Info(msg)
			r.event(ctx, kluctlDeployment, source.GetArtifact().Revision, events.EventSeverityInfo, msg, nil)
			r.recordReadiness(ctx, kluctlDeployment)
			return ctrl.Result{RequeueAfter: r.requeueDependency}, nil
		}
		log.Info("All dependencies are ready, proceeding with reconciliation")
	}

	// record reconciliation duration
	if r.MetricsRecorder != nil {
		objRef, err := reference.GetReference(r.Scheme, &kluctlDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		defer r.MetricsRecorder.RecordDuration(*objRef, reconcileStart)
	}

	// set the reconciliation status to progressing
	kluctlDeployment = kluctlv1.KluctlDeploymentProgressing(kluctlDeployment, "reconciliation in progress")
	if err := r.patchStatus(ctx, req, kluctlDeployment.Status); err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.recordReadiness(ctx, kluctlDeployment)

	// reconcile kluctlDeployment by applying the latest revision
	reconciledKluctlDeployment, reconcileErr := r.reconcile(ctx, *kluctlDeployment.DeepCopy(), source)
	if err := r.patchStatus(ctx, req, reconciledKluctlDeployment.Status); err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.recordReadiness(ctx, reconciledKluctlDeployment)

	// broadcast the reconciliation failure and requeue at the specified retry interval
	if reconcileErr != nil {
		log.Error(reconcileErr, fmt.Sprintf("Reconciliation failed after %s, next try in %s",
			time.Since(reconcileStart).String(),
			kluctlDeployment.GetRetryInterval().String()),
			"revision",
			source.GetArtifact().Revision)
		r.event(ctx, reconciledKluctlDeployment, source.GetArtifact().Revision, events.EventSeverityError,
			reconcileErr.Error(), nil)
		return ctrl.Result{RequeueAfter: kluctlDeployment.GetRetryInterval()}, nil
	}

	// broadcast the reconciliation result and requeue at the specified interval
	msg := fmt.Sprintf("Reconciliation finished in %s, next run in %s",
		time.Since(reconcileStart).String(),
		kluctlDeployment.Spec.Interval.Duration.String())
	log.Info(msg, "revision", source.GetArtifact().Revision)
	r.event(ctx, reconciledKluctlDeployment, source.GetArtifact().Revision, events.EventSeverityInfo,
		msg, map[string]string{kluctlv1.GroupVersion.Group + "/commit_status": "update"})
	return ctrl.Result{RequeueAfter: kluctlDeployment.Spec.Interval.Duration}, nil
}

func (r *KluctlDeploymentReconciler) reconcile(
	ctx context.Context,
	kluctlDeployment kluctlv1.KluctlDeployment,
	source sourcev1.Source) (kluctlv1.KluctlDeployment, error) {
	// record the value of the reconciliation request, if any
	if v, ok := meta.ReconcileAnnotationValue(kluctlDeployment.GetAnnotations()); ok {
		kluctlDeployment.Status.SetLastHandledReconcileRequest(v)
	}

	pp, err := prepareProject(ctx, r, kluctlDeployment, source)
	if err != nil {
		return kluctlv1.KluctlDeploymentNotReady(
			kluctlDeployment,
			pp.source.GetArtifact().Revision,
			pp.targetHash,
			kluctlv1.PrepareFailedReason,
			err.Error(),
		), err
	}
	defer pp.cleanup()

	kluctlDeployment.Status.InvolvedRepos = nil
	for u, repo := range pp.metadata.InvolvedRepos {
		ir2 := kluctlv1.InvolvedRepo{
			URL: u,
		}

		for _, ir := range repo {
			ir2.Patterns = append(ir2.Patterns, kluctlv1.InvolvedRepoPattern{
				Pattern: ir.RefPattern,
				Refs:    ir.Refs,
			})
		}

		kluctlDeployment.Status.InvolvedRepos = append(kluctlDeployment.Status.InvolvedRepos, ir2)
	}

	if kluctlDeployment.Status.LastAttemptedTarget != nil && kluctlDeployment.Status.LastAttemptedTarget.TargetHash == pp.targetHash {
		if kluctlDeployment.Status.LastDeployedTarget != nil && kluctlDeployment.Status.LastDeployedTarget.TargetHash == pp.targetHash {
			return kluctlv1.KluctlDeploymentReady(
				kluctlDeployment,
				pp.source.GetArtifact().Revision,
				pp.targetHash,
				kluctlv1.ReconciliationSkippedReason,
				fmt.Sprintf("Skipped revision as target did not change: %s", pp.targetHash),
			), nil
		} else {
			return kluctlv1.KluctlDeploymentNotReady(
				kluctlDeployment,
				pp.source.GetArtifact().Revision,
				pp.targetHash,
				kluctlv1.ReconciliationSkippedReason,
				fmt.Sprintf("Skipped revision as target did not change: %s", pp.targetHash),
			), nil
		}
	}

	// deploy the kluctl project
	deployResult, err := pp.kluctlDeploy(ctx)
	kluctlDeployment.Status.LastDeployResult = deployResult
	if err != nil {
		return kluctlv1.KluctlDeploymentNotReady(
			kluctlDeployment,
			pp.source.GetArtifact().Revision,
			pp.targetHash,
			kluctlv1.DeployFailedReason,
			err.Error(),
		), err
	}

	// run garbage collection for stale objects that do not have pruning disabled
	pruneResult, err := pp.kluctlPrune(ctx)
	kluctlDeployment.Status.LastPruneResult = pruneResult
	if err != nil {
		return kluctlv1.KluctlDeploymentNotReady(
			kluctlDeployment,
			pp.source.GetArtifact().Revision,
			pp.targetHash,
			kluctlv1.PruneFailedReason,
			err.Error(),
		), err
	}

	return kluctlv1.KluctlDeploymentReady(
		kluctlDeployment,
		pp.source.GetArtifact().Revision,
		pp.targetHash,
		kluctlv1.ReconciliationSucceededReason,
		fmt.Sprintf("Deployed revision: %s", pp.source.GetArtifact().Revision),
	), nil
}

func (r *KluctlDeploymentReconciler) checkDependencies(source sourcev1.Source, kluctlDeployment kluctlv1.KluctlDeployment) error {
	for _, d := range kluctlDeployment.Spec.DependsOn {
		if d.Namespace == "" {
			d.Namespace = kluctlDeployment.GetNamespace()
		}
		dName := types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Name,
		}
		var k kluctlv1.KluctlDeployment
		err := r.Get(context.Background(), dName, &k)
		if err != nil {
			return fmt.Errorf("unable to get '%s' dependency: %w", dName, err)
		}

		if len(k.Status.Conditions) == 0 || k.Generation != k.Status.ObservedGeneration {
			return fmt.Errorf("dependency '%s' is not ready", dName)
		}

		if !apimeta.IsStatusConditionTrue(k.Status.Conditions, meta.ReadyCondition) {
			return fmt.Errorf("dependency '%s' is not ready", dName)
		}

		if k.Spec.SourceRef.Name == kluctlDeployment.Spec.SourceRef.Name && k.Spec.SourceRef.Namespace == kluctlDeployment.Spec.SourceRef.Namespace && k.Spec.SourceRef.Kind == kluctlDeployment.Spec.SourceRef.Kind && source.GetArtifact().Revision != k.Status.LastDeployedRevision {
			return fmt.Errorf("dependency '%s' is not updated yet", dName)
		}
	}

	return nil
}

func (r *KluctlDeploymentReconciler) download(source sourcev1.Source, tmpDir string) error {
	artifact := source.GetArtifact()
	artifactURL := artifact.URL
	if hostname := os.Getenv("SOURCE_CONTROLLER_LOCALHOST"); hostname != "" {
		u, err := url.Parse(artifactURL)
		if err != nil {
			return err
		}
		u.Host = hostname
		artifactURL = u.String()
	}

	req, err := retryablehttp.NewRequest(http.MethodGet, artifactURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create a new request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download artifact, error: %w", err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download artifact from %s, status: %s", artifactURL, resp.Status)
	}

	var buf bytes.Buffer

	// verify checksum matches origin
	if err := r.verifyArtifact(artifact, &buf, resp.Body); err != nil {
		return err
	}

	// extract
	if _, err = untar.Untar(&buf, filepath.Join(tmpDir, "source")); err != nil {
		return fmt.Errorf("failed to untar artifact, error: %w", err)
	}

	return nil
}

func (r *KluctlDeploymentReconciler) verifyArtifact(artifact *sourcev1.Artifact, buf *bytes.Buffer, reader io.Reader) error {
	hasher := sha256.New()

	// for backwards compatibility with source-controller v0.17.2 and older
	if len(artifact.Checksum) == 40 {
		hasher = sha1.New()
	}

	// compute checksum
	mw := io.MultiWriter(hasher, buf)
	if _, err := io.Copy(mw, reader); err != nil {
		return err
	}

	if checksum := fmt.Sprintf("%x", hasher.Sum(nil)); checksum != artifact.Checksum {
		return fmt.Errorf("failed to verify artifact: computed checksum '%s' doesn't match advertised '%s'",
			checksum, artifact.Checksum)
	}

	return nil
}

func (r *KluctlDeploymentReconciler) getSource(ctx context.Context, kluctlDeployment kluctlv1.KluctlDeployment) (sourcev1.Source, error) {
	var source sourcev1.Source
	sourceNamespace := kluctlDeployment.GetNamespace()
	if kluctlDeployment.Spec.SourceRef.Namespace != "" {
		sourceNamespace = kluctlDeployment.Spec.SourceRef.Namespace
	}
	namespacedName := types.NamespacedName{
		Namespace: sourceNamespace,
		Name:      kluctlDeployment.Spec.SourceRef.Name,
	}

	if r.NoCrossNamespaceRefs && sourceNamespace != kluctlDeployment.GetNamespace() {
		return source, acl.AccessDeniedError(
			fmt.Sprintf("can't access '%s/%s', cross-namespace references have been blocked",
				kluctlDeployment.Spec.SourceRef.Kind, namespacedName))
	}

	switch kluctlDeployment.Spec.SourceRef.Kind {
	case sourcev1.GitRepositoryKind:
		var repository sourcev1.GitRepository
		err := r.Client.Get(ctx, namespacedName, &repository)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return source, err
			}
			return source, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		source = &repository
	case sourcev1.BucketKind:
		var bucket sourcev1.Bucket
		err := r.Client.Get(ctx, namespacedName, &bucket)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return source, err
			}
			return source, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		source = &bucket
	default:
		return source, fmt.Errorf("source `%s` kind '%s' not supported",
			kluctlDeployment.Spec.SourceRef.Name, kluctlDeployment.Spec.SourceRef.Kind)
	}
	return source, nil
}

func (r *KluctlDeploymentReconciler) finalize(ctx context.Context, kluctlDeployment kluctlv1.KluctlDeployment) (ctrl.Result, error) {
	if kluctlDeployment.Spec.Prune &&
		!kluctlDeployment.Spec.Suspend {

		source, err := r.getSource(ctx, kluctlDeployment)
		if err == nil {
			pp, err := prepareProject(ctx, r, kluctlDeployment, source)
			if err == nil {
				defer pp.cleanup()
				_, _ = pp.kluctlDelete(ctx)
			}
		}
	}

	// Record deleted status
	r.recordReadiness(ctx, kluctlDeployment)

	// Remove our finalizer from the list and update it
	controllerutil.RemoveFinalizer(&kluctlDeployment, kluctlv1.KluctlDeploymentFinalizer)
	if err := r.Update(ctx, &kluctlDeployment, client.FieldOwner(r.statusManager)); err != nil {
		return ctrl.Result{}, err
	}

	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}

func (r *KluctlDeploymentReconciler) event(ctx context.Context, kluctlDeployment kluctlv1.KluctlDeployment, revision, severity, msg string, metadata map[string]string) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	if revision != "" {
		metadata[kluctlv1.GroupVersion.Group+"/revision"] = revision
	}

	reason := severity
	if c := apimeta.FindStatusCondition(kluctlDeployment.Status.Conditions, meta.ReadyCondition); c != nil {
		reason = c.Reason
	}

	eventtype := "Normal"
	if severity == events.EventSeverityError {
		eventtype = "Warning"
	}

	r.EventRecorder.AnnotatedEventf(&kluctlDeployment, metadata, eventtype, reason, msg)
}

func (r *KluctlDeploymentReconciler) recordReadiness(ctx context.Context, kluctlDeployment kluctlv1.KluctlDeployment) {
	if r.MetricsRecorder == nil {
		return
	}
	log := ctrl.LoggerFrom(ctx)

	objRef, err := reference.GetReference(r.Scheme, &kluctlDeployment)
	if err != nil {
		log.Error(err, "unable to record readiness metric")
		return
	}
	if rc := apimeta.FindStatusCondition(kluctlDeployment.Status.Conditions, meta.ReadyCondition); rc != nil {
		r.MetricsRecorder.RecordCondition(*objRef, *rc, !kluctlDeployment.DeletionTimestamp.IsZero())
	} else {
		r.MetricsRecorder.RecordCondition(*objRef, metav1.Condition{
			Type:   meta.ReadyCondition,
			Status: metav1.ConditionUnknown,
		}, !kluctlDeployment.DeletionTimestamp.IsZero())
	}
}

func (r *KluctlDeploymentReconciler) recordSuspension(ctx context.Context, kluctlDeployment kluctlv1.KluctlDeployment) {
	if r.MetricsRecorder == nil {
		return
	}
	log := ctrl.LoggerFrom(ctx)

	objRef, err := reference.GetReference(r.Scheme, &kluctlDeployment)
	if err != nil {
		log.Error(err, "unable to record suspended metric")
		return
	}

	if !kluctlDeployment.DeletionTimestamp.IsZero() {
		r.MetricsRecorder.RecordSuspend(*objRef, false)
	} else {
		r.MetricsRecorder.RecordSuspend(*objRef, kluctlDeployment.Spec.Suspend)
	}
}

func (r *KluctlDeploymentReconciler) patchStatus(ctx context.Context, req ctrl.Request, newStatus kluctlv1.KluctlDeploymentStatus) error {
	var kluctlDeployment kluctlv1.KluctlDeployment
	if err := r.Get(ctx, req.NamespacedName, &kluctlDeployment); err != nil {
		return err
	}

	patch := client.MergeFrom(kluctlDeployment.DeepCopy())
	kluctlDeployment.Status = newStatus
	return r.Status().Patch(ctx, &kluctlDeployment, patch, client.FieldOwner(r.statusManager))
}
