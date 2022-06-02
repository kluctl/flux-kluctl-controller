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
	"github.com/kluctl/kluctl/v2/pkg/status"
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

// KluctlProjectReconciler reconciles a KluctlDeployment object
type KluctlProjectReconciler struct {
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

	Impl KluctlProjectReconcilerImpl
}

// KluctlProjectReconcilerOptions contains options for the KluctlProjectReconciler.
type KluctlProjectReconcilerOptions struct {
	MaxConcurrentReconciles   int
	HTTPRetry                 int
	DependencyRequeueInterval time.Duration
}

type KluctlProjectReconcilerImpl interface {
	NewObject() KluctlProjectHolder
	NewObjectList() KluctlProjectListHolder
	Reconcile(ctx context.Context, obj KluctlProjectHolder, source sourcev1.Source) error
	Finalize(ctx context.Context, obj KluctlProjectHolder)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KluctlProjectReconciler) SetupWithManager(mgr ctrl.Manager, opts KluctlProjectReconcilerOptions) error {
	const (
		gitRepositoryIndexKey string = ".metadata.gitRepository"
		bucketIndexKey        string = ".metadata.bucket"
	)

	// Index the KluctlDeployments by the GitRepository references they (may) point at.
	if err := mgr.GetCache().IndexField(context.TODO(), r.Impl.NewObject(), gitRepositoryIndexKey,
		r.indexBy(sourcev1.GitRepositoryKind)); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	// Index the KluctlDeployments by the Bucket references they (may) point at.
	if err := mgr.GetCache().IndexField(context.TODO(), r.Impl.NewObject(), bucketIndexKey,
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
		For(r.Impl.NewObject(), builder.WithPredicates(
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

func (r *KluctlProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reconcileStart := time.Now()

	ctx = status.NewContext(ctx, status.NewSimpleStatusHandler(func(message string) {
		log.Info(message)
	}, false))

	obj := r.Impl.NewObject()
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	projectSpec := obj.GetKluctlProject()
	timingSpec := obj.GetKluctlTiming()

	ctx, cancel := context.WithTimeout(ctx, timingSpec.GetTimeout())
	defer cancel()

	// Record suspended status metric
	defer r.recordSuspension(ctx, obj)

	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(obj, kluctlv1.KluctlDeploymentFinalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(KluctlProjectHolder))
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
	if projectSpec.Suspend {
		log.Info("Reconciliation is suspended for this object")
		return ctrl.Result{}, nil
	}

	// resolve source reference
	source, err := r.getSource(ctx, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			msg := fmt.Sprintf("Source '%s' not found", projectSpec.SourceRef.String())
			kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.ArtifactFailedReason, msg, obj.GetGeneration(), "")
			if err := r.patchProjectStatus(ctx, req, *obj.GetKluctlStatus()); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			r.recordReadiness(ctx, obj)
			log.Info(msg)
			// do not requeue immediately, when the source is created the watcher should trigger a reconciliation
			return ctrl.Result{RequeueAfter: timingSpec.GetRetryInterval()}, nil
		}

		if acl.IsAccessDenied(err) {
			kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, apiacl.AccessDeniedReason, err.Error(), obj.GetGeneration(), "")
			if err := r.patchProjectStatus(ctx, req, *obj.GetKluctlStatus()); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			log.Error(err, "access denied to cross-namespace source")
			r.recordReadiness(ctx, obj)
			r.event(ctx, obj, "unknown", events.EventSeverityError, err.Error(), nil)
			return ctrl.Result{RequeueAfter: timingSpec.GetRetryInterval()}, nil
		}

		// retry on transient errors
		return ctrl.Result{Requeue: true}, err
	}

	if source.GetArtifact() == nil {
		msg := "Source is not ready, artifact not found"
		kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.ArtifactFailedReason, msg, obj.GetGeneration(), "")
		if err := r.patchProjectStatus(ctx, req, *obj.GetKluctlStatus()); err != nil {
			log.Error(err, "unable to update status for artifact not found")
			return ctrl.Result{Requeue: true}, err
		}
		r.recordReadiness(ctx, obj)
		log.Info(msg)
		// do not requeue immediately, when the artifact is created the watcher should trigger a reconciliation
		return ctrl.Result{RequeueAfter: timingSpec.GetRetryInterval()}, nil
	}

	// check dependencies
	if len(projectSpec.DependsOn) > 0 {
		if err := r.checkDependencies(source, obj); err != nil {
			kluctlv1.SetKluctlProjectReadiness(obj.GetKluctlStatus(), metav1.ConditionFalse, kluctlv1.DependencyNotReadyReason, err.Error(), obj.GetGeneration(), source.GetArtifact().Revision)
			if err := r.patchProjectStatus(ctx, req, *obj.GetKluctlStatus()); err != nil {
				log.Error(err, "unable to update status for dependency not ready")
				return ctrl.Result{Requeue: true}, err
			}
			// we can't rely on exponential backoff because it will prolong the execution too much,
			// instead we requeue on a fix interval.
			msg := fmt.Sprintf("Dependencies do not meet ready condition, retrying in %s", r.requeueDependency.String())
			log.Info(msg)
			r.event(ctx, obj, source.GetArtifact().Revision, events.EventSeverityInfo, msg, nil)
			r.recordReadiness(ctx, obj)
			return ctrl.Result{RequeueAfter: r.requeueDependency}, nil
		}
		log.Info("All dependencies are ready, proceeding with reconciliation")
	}

	// record reconciliation duration
	if r.MetricsRecorder != nil {
		objRef, err := reference.GetReference(r.Scheme, obj)
		if err != nil {
			return ctrl.Result{}, err
		}
		defer r.MetricsRecorder.RecordDuration(*objRef, reconcileStart)
	}

	// set the reconciliation status to progressing
	kluctlv1.KluctlProjectProgressing(obj.GetKluctlStatus(), "reconciliation in progress")
	if err := r.patchProjectStatus(ctx, req, *obj.GetKluctlStatus()); err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.recordReadiness(ctx, obj)

	// record the value of the reconciliation request, if any
	if v, ok := meta.ReconcileAnnotationValue(obj.GetAnnotations()); ok {
		obj.GetKluctlStatus().SetLastHandledReconcileRequest(v)
	}

	// reconcile kluctlDeployment by applying the latest revision
	reconcileErr := r.Impl.Reconcile(ctx, obj, source)
	if err := r.patchFullStatus(ctx, req, obj.GetFullStatus()); err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	r.recordReadiness(ctx, obj)

	// broadcast the reconciliation failure and requeue at the specified retry interval
	if reconcileErr != nil {
		log.Error(reconcileErr, fmt.Sprintf("Reconciliation failed after %s, next try in %s",
			time.Since(reconcileStart).String(),
			timingSpec.GetRetryInterval().String()),
			"revision",
			source.GetArtifact().Revision)
		r.event(ctx, obj, source.GetArtifact().Revision, events.EventSeverityError,
			reconcileErr.Error(), nil)
		return ctrl.Result{RequeueAfter: timingSpec.GetRetryInterval()}, nil
	}

	// broadcast the reconciliation result and requeue at the specified interval
	msg := fmt.Sprintf("Reconciliation finished in %s, next run in %s",
		time.Since(reconcileStart).String(),
		timingSpec.Interval.Duration.String())
	log.Info(msg, "revision", source.GetArtifact().Revision)
	r.event(ctx, obj, source.GetArtifact().Revision, events.EventSeverityInfo,
		msg, map[string]string{kluctlv1.GroupVersion.Group + "/commit_status": "update"})
	return ctrl.Result{RequeueAfter: timingSpec.Interval.Duration}, nil
}

func (r *KluctlProjectReconciler) checkDependencies(source sourcev1.Source, obj KluctlProjectHolder) error {
	projectSpec := obj.GetKluctlProject()

	for _, d := range projectSpec.DependsOn {
		if d.Namespace == "" {
			d.Namespace = obj.GetNamespace()
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

		if k.Spec.SourceRef.Name == projectSpec.SourceRef.Name &&
			k.Spec.SourceRef.Namespace == projectSpec.SourceRef.Namespace &&
			k.Spec.SourceRef.Kind == projectSpec.SourceRef.Kind &&
			k.Status.LastSuccessfulReconcile != nil &&
			source.GetArtifact().Revision != k.Status.LastSuccessfulReconcile.Revision {
			return fmt.Errorf("dependency '%s' is not updated yet", dName)
		}
	}

	return nil
}

func (r *KluctlProjectReconciler) download(source sourcev1.Source, tmpDir string) error {
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

func (r *KluctlProjectReconciler) verifyArtifact(artifact *sourcev1.Artifact, buf *bytes.Buffer, reader io.Reader) error {
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

func (r *KluctlProjectReconciler) getSource(ctx context.Context, obj KluctlProjectHolder) (sourcev1.Source, error) {
	projectSpec := obj.GetKluctlProject()

	var source sourcev1.Source
	sourceNamespace := obj.GetNamespace()
	if projectSpec.SourceRef.Namespace != "" {
		sourceNamespace = projectSpec.SourceRef.Namespace
	}
	namespacedName := types.NamespacedName{
		Namespace: sourceNamespace,
		Name:      projectSpec.SourceRef.Name,
	}

	if r.NoCrossNamespaceRefs && sourceNamespace != obj.GetNamespace() {
		return source, acl.AccessDeniedError(
			fmt.Sprintf("can't access '%s/%s', cross-namespace references have been blocked",
				projectSpec.SourceRef.Kind, namespacedName))
	}

	switch projectSpec.SourceRef.Kind {
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
			projectSpec.SourceRef.Name, projectSpec.SourceRef.Kind)
	}
	return source, nil
}

func (r *KluctlProjectReconciler) doFinalize(ctx context.Context, obj KluctlProjectHolder) {
	panic("not implemented")
}

func (r *KluctlProjectReconciler) finalize(ctx context.Context, obj KluctlProjectHolder) (ctrl.Result, error) {
	r.Impl.Finalize(ctx, obj)

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

func (r *KluctlProjectReconciler) event(ctx context.Context, obj KluctlProjectHolder, revision, severity, msg string, metadata map[string]string) {
	if metadata == nil {
		metadata = map[string]string{}
	}
	if revision != "" {
		metadata[kluctlv1.GroupVersion.Group+"/revision"] = revision
	}

	reason := severity
	if c := apimeta.FindStatusCondition(obj.GetConditions(), meta.ReadyCondition); c != nil {
		reason = c.Reason
	}

	eventtype := "Normal"
	if severity == events.EventSeverityError {
		eventtype = "Warning"
	}

	r.EventRecorder.AnnotatedEventf(obj, metadata, eventtype, reason, msg)
}

func (r *KluctlProjectReconciler) recordReadiness(ctx context.Context, obj KluctlProjectHolder) {
	if r.MetricsRecorder == nil {
		return
	}
	log := ctrl.LoggerFrom(ctx)

	objRef, err := reference.GetReference(r.Scheme, obj)
	if err != nil {
		log.Error(err, "unable to record readiness metric")
		return
	}
	if rc := apimeta.FindStatusCondition(obj.GetConditions(), meta.ReadyCondition); rc != nil {
		r.MetricsRecorder.RecordCondition(*objRef, *rc, !obj.GetDeletionTimestamp().IsZero())
	} else {
		r.MetricsRecorder.RecordCondition(*objRef, metav1.Condition{
			Type:   meta.ReadyCondition,
			Status: metav1.ConditionUnknown,
		}, !obj.GetDeletionTimestamp().IsZero())
	}
}

func (r *KluctlProjectReconciler) recordSuspension(ctx context.Context, obj KluctlProjectHolder) {
	if r.MetricsRecorder == nil {
		return
	}
	log := ctrl.LoggerFrom(ctx)

	objRef, err := reference.GetReference(r.Scheme, obj)
	if err != nil {
		log.Error(err, "unable to record suspended metric")
		return
	}

	if !obj.GetDeletionTimestamp().IsZero() {
		r.MetricsRecorder.RecordSuspend(*objRef, false)
	} else {
		r.MetricsRecorder.RecordSuspend(*objRef, obj.GetKluctlProject().Suspend)
	}
}

func (r *KluctlProjectReconciler) patchProjectStatus(ctx context.Context, req ctrl.Request, newStatus kluctlv1.KluctlProjectStatus) error {
	obj := r.Impl.NewObject()
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return err
	}

	patch := client.MergeFrom(obj.DeepCopyObject().(KluctlProjectHolder))
	*obj.GetKluctlStatus() = newStatus
	return r.Status().Patch(ctx, obj, patch, client.FieldOwner(r.statusManager))
}

func (r *KluctlProjectReconciler) patchFullStatus(ctx context.Context, req ctrl.Request, newStatus any) error {
	obj := r.Impl.NewObject()
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return err
	}

	patch := client.MergeFrom(obj.DeepCopyObject().(KluctlProjectHolder))
	obj.SetFullStatus(newStatus)
	return r.Status().Patch(ctx, obj, patch, client.FieldOwner(r.statusManager))
}
