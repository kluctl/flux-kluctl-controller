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
	"context"
	"errors"
	"fmt"
	"github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/flux-kluctl-controller/controllers/source-controller"
	types2 "github.com/kluctl/kluctl/pkg/types"
	"github.com/kluctl/kluctl/pkg/yaml"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fluxcd/pkg/runtime/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	helper "github.com/fluxcd/pkg/runtime/controller"
	"github.com/fluxcd/pkg/runtime/events"
	"github.com/fluxcd/pkg/runtime/patch"
	"github.com/fluxcd/pkg/runtime/predicates"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	serror "github.com/kluctl/flux-kluctl-controller/controllers/source-controller/error"
	sreconcile "github.com/kluctl/flux-kluctl-controller/controllers/source-controller/reconcile"
	"github.com/kluctl/flux-kluctl-controller/controllers/source-controller/reconcile/summarize"
	"github.com/kluctl/flux-kluctl-controller/controllers/source-controller/util"
)

// kluctlProjectReadyCondition contains the information required to summarize a
// v1beta2.KluctlProject Ready Condition.
var kluctlProjectReadyCondition = summarize.Conditions{
	Target: meta.ReadyCondition,
	Owned: []string{
		sourcev1.StorageOperationFailedCondition,
		sourcev1.ArtifactOutdatedCondition,
		sourcev1.ArtifactInStorageCondition,
		v1alpha1.ArchiveFailedCondition,
		meta.ReadyCondition,
		meta.ReconcilingCondition,
		meta.StalledCondition,
	},
	Summarize: []string{
		sourcev1.StorageOperationFailedCondition,
		sourcev1.ArtifactOutdatedCondition,
		sourcev1.ArtifactInStorageCondition,
		v1alpha1.ArchiveFailedCondition,
		meta.StalledCondition,
		meta.ReconcilingCondition,
	},
	NegativePolarity: []string{
		sourcev1.StorageOperationFailedCondition,
		sourcev1.ArtifactOutdatedCondition,
		v1alpha1.ArchiveFailedCondition,
		meta.StalledCondition,
		meta.ReconcilingCondition,
	},
}

// kluctlProjectFailConditions contains the conditions that represent a failure.
var kluctlProjectFailConditions = []string{
	sourcev1.StorageOperationFailedCondition,
	v1alpha1.ArchiveFailedCondition,
}

// +kubebuilder:rbac:groups=kluctl.io,resources=kluctlprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kluctl.io,resources=kluctlprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kluctl.io,resources=kluctlprojects/finalizers,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// KluctlProjectReconciler reconciles a v1beta2.KluctlProject object.
type KluctlProjectReconciler struct {
	client.Client
	kuberecorder.EventRecorder
	helper.Metrics

	Storage        *source_controller.Storage
	ControllerName string

	requeueDependency time.Duration
}

type KluctlProjectReconcilerOptions struct {
	MaxConcurrentReconciles   int
	DependencyRequeueInterval time.Duration
	RateLimiter               ratelimiter.RateLimiter
}

// kluctlProjectReconcileFunc is the function type for all the
// v1beta2.KluctlProject (sub)reconcile functions.
type kluctlProjectReconcileFunc func(ctx context.Context, obj *v1alpha1.KluctlProject, archiveInfo *v1alpha1.ArchiveInfo, metadata *types2.ArchiveMetadata, dir string) (sreconcile.Result, error)

func (r *KluctlProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return r.SetupWithManagerAndOptions(mgr, KluctlProjectReconcilerOptions{})
}

func (r *KluctlProjectReconciler) SetupWithManagerAndOptions(mgr ctrl.Manager, opts KluctlProjectReconcilerOptions) error {
	r.requeueDependency = opts.DependencyRequeueInterval

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.KluctlProject{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicates.ReconcileRequestedPredicate{}),
		)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opts.MaxConcurrentReconciles,
			RateLimiter:             opts.RateLimiter,
		}).
		Complete(r)
}

func (r *KluctlProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	start := time.Now()
	log := ctrl.LoggerFrom(ctx)

	// Fetch the KluctlProject
	obj := &v1alpha1.KluctlProject{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Record suspended status metric
	r.RecordSuspend(ctx, obj, obj.Spec.Suspend)

	// Return early if the object is suspended
	if obj.Spec.Suspend {
		log.Info("reconciliation is suspended for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper with the current version of the object.
	patchHelper, err := patch.NewHelper(obj, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// recResult stores the abstracted reconcile result.
	var recResult sreconcile.Result

	// Always attempt to patch the object and status after each reconciliation
	// NOTE: The final runtime result and error are set in this block.
	defer func() {
		summarizeHelper := summarize.NewHelper(r.EventRecorder, patchHelper)
		summarizeOpts := []summarize.Option{
			summarize.WithConditions(kluctlProjectReadyCondition),
			summarize.WithReconcileResult(recResult),
			summarize.WithReconcileError(retErr),
			summarize.WithIgnoreNotFound(),
			summarize.WithProcessors(
				summarize.RecordContextualError,
				summarize.RecordReconcileReq,
			),
			summarize.WithResultBuilder(sreconcile.AlwaysRequeueResultBuilder{RequeueAfter: obj.GetRequeueAfter()}),
			summarize.WithPatchFieldOwner(r.ControllerName),
		}
		result, retErr = summarizeHelper.SummarizeAndPatch(ctx, obj, summarizeOpts...)

		// Always record readiness and duration metrics
		r.Metrics.RecordReadiness(ctx, obj)
		r.Metrics.RecordDuration(ctx, obj, start)
	}()

	// Add finalizer first if not exist to avoid the race condition
	// between init and delete
	if !controllerutil.ContainsFinalizer(obj, sourcev1.SourceFinalizer) {
		controllerutil.AddFinalizer(obj, sourcev1.SourceFinalizer)
		recResult = sreconcile.ResultRequeue
		return
	}

	// Examine if the object is under deletion
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		recResult, retErr = r.reconcileDelete(ctx, obj)
		return
	}

	// Reconcile actual object
	reconcilers := []kluctlProjectReconcileFunc{
		r.reconcileStorage,
		r.reconcileArchive,
		r.reconcileArtifact,
	}
	recResult, retErr = r.reconcile(ctx, obj, reconcilers)
	return
}

// reconcile iterates through the kluctlProjectReconcileFunc tasks for the
// object. It returns early on the first call that returns
// reconcile.ResultRequeue, or produces an error.
func (r *KluctlProjectReconciler) reconcile(ctx context.Context, obj *v1alpha1.KluctlProject, reconcilers []kluctlProjectReconcileFunc) (sreconcile.Result, error) {
	oldObj := obj.DeepCopy()

	// Mark as reconciling if generation differs
	if obj.Generation != obj.Status.ObservedGeneration {
		conditions.MarkReconciling(obj, "NewGeneration", "reconciling new object generation (%d)", obj.Generation)
	}

	// Create temp dir for kluctl archive
	tmpDir, err := util.TempDirForObj("", obj)
	if err != nil {
		e := &serror.Event{
			Err:    fmt.Errorf("failed to create temporary working directory: %w", err),
			Reason: sourcev1.DirCreationFailedReason,
		}
		conditions.MarkTrue(obj, sourcev1.StorageOperationFailedCondition, e.Reason, e.Err.Error())
		return sreconcile.ResultEmpty, e
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			ctrl.LoggerFrom(ctx).Error(err, "failed to remove temporary working directory")
		}
	}()
	conditions.Delete(obj, sourcev1.StorageOperationFailedCondition)

	// Run the sub-reconcilers and build the result of reconciliation.
	var (
		archiveInfo v1alpha1.ArchiveInfo
		metadata types2.ArchiveMetadata

		res    sreconcile.Result
		resErr error
	)
	for _, rec := range reconcilers {
		recResult, err := rec(ctx, obj, &archiveInfo, &metadata, tmpDir)
		// Exit immediately on ResultRequeue.
		if recResult == sreconcile.ResultRequeue {
			return sreconcile.ResultRequeue, nil
		}
		// If an error is received, prioritize the returned results because an
		// error also means immediate requeue.
		if err != nil {
			resErr = err
			res = recResult
			break
		}
		// Prioritize requeue request in the result.
		res = sreconcile.LowestRequeuingResult(res, recResult)
	}

	r.notify(oldObj, obj, archiveInfo, res, resErr)

	return res, resErr
}

// notify emits notification related to the reconciliation.
func (r *KluctlProjectReconciler) notify(oldObj, newObj *v1alpha1.KluctlProject, archiveInfo v1alpha1.ArchiveInfo, res sreconcile.Result, resErr error) {
	// Notify successful reconciliation for new artifact and recovery from any
	// failure.
	if resErr == nil && res == sreconcile.ResultSuccess && newObj.Status.Artifact != nil {
		annotations := map[string]string{
			sourcev1.GroupVersion.Group + "/revision": newObj.Status.Artifact.Revision,
			sourcev1.GroupVersion.Group + "/checksum": newObj.Status.Artifact.Checksum,
		}

		var oldChecksum string
		if oldObj.GetArtifact() != nil {
			oldChecksum = oldObj.GetArtifact().Checksum
		}

		message := fmt.Sprintf("stored artifact with revision '%s'", archiveInfo.String())

		// Notify on new artifact and failure recovery.
		if oldChecksum != newObj.GetArtifact().Checksum {
			r.AnnotatedEventf(newObj, annotations, corev1.EventTypeNormal,
				"NewArtifact", message)
		} else {
			if sreconcile.FailureRecovery(oldObj, newObj, kluctlProjectFailConditions) {
				r.AnnotatedEventf(newObj, annotations, corev1.EventTypeNormal,
					meta.SucceededReason, message)
			}
		}
	}
}

// reconcileStorage ensures the current state of the storage matches the
// desired and previously observed state.
//
// All Artifacts for the object except for the current one in the Status are
// garbage collected from the Storage.
// If the Artifact in the Status of the object disappeared from the Storage,
// it is removed from the object.
// If the object does not have an Artifact in its Status, a Reconciling
// condition is added.
// The hostname of any URL in the Status of the object are updated, to ensure
// they match the Storage server hostname of current runtime.
func (r *KluctlProjectReconciler) reconcileStorage(ctx context.Context,
	obj *v1alpha1.KluctlProject, _ *v1alpha1.ArchiveInfo, _ *types2.ArchiveMetadata, _ string) (sreconcile.Result, error) {
	// Garbage collect previous advertised artifact(s) from storage
	_ = r.garbageCollect(ctx, obj)

	// Determine if the advertised artifact is still in storage
	if artifact := obj.GetArtifact(); artifact != nil && !r.Storage.ArtifactExist(*artifact) {
		obj.Status.Artifact = nil
		obj.Status.URL = ""
		// Remove the condition as the artifact doesn't exist.
		conditions.Delete(obj, sourcev1.ArtifactInStorageCondition)
	}

	// Record that we do not have an artifact
	if obj.GetArtifact() == nil {
		conditions.MarkReconciling(obj, "NoArtifact", "no artifact for resource in storage")
		conditions.Delete(obj, sourcev1.ArtifactInStorageCondition)
		return sreconcile.ResultSuccess, nil
	}

	// Always update URLs to ensure hostname is up-to-date
	// TODO(hidde): we may want to send out an event only if we notice the URL has changed
	r.Storage.SetArtifactURL(obj.GetArtifact())
	obj.Status.URL = r.Storage.SetHostname(obj.Status.URL)

	return sreconcile.ResultSuccess, nil
}

// reconcileArchive ensures the kluctl archive can be created. This includes
// cloning and checking out all involved git repositories.
//
// We use the "kluctl archive" command and rely on it to fully handle all involved/dependent git repositories.
// Git credentials are loaded from a secret and passed via environment variables.
//
// The archive command can be run in two flavors. One with the metadata.yml included into the archive and one with
// the metadata.yml stored externally. We use the second form to avoid false-positive reconciliations of all targets
// when only one changed. This is because the archive.tar.gz might stay unchanged when target configuration changes
// through external target configuration git repositories.
func (r *KluctlProjectReconciler) reconcileArchive(ctx context.Context,
	obj *v1alpha1.KluctlProject, archiveInfo *v1alpha1.ArchiveInfo, metadata *types2.ArchiveMetadata, dir string) (sreconcile.Result, error) {

	cmd := kluctlCaller{
		workDir: dir,
	}
	defer cmd.deleteTmpFiles()

	handleError := func(err error, reason string) (sreconcile.Result, error) {
		e := &serror.Event{
			Err:    err,
			Reason: reason,
		}
		conditions.MarkTrue(obj, v1alpha1.ArchiveFailedCondition, e.Reason, e.Err.Error())
		// Return error as the world as observed may change
		return sreconcile.ResultEmpty, e
	}

	var err error
	if obj.Spec.SecretRef != nil {
		// Attempt to retrieve secret
		name := types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.Spec.SecretRef.Name,
		}
		var secret corev1.Secret
		if err := r.Client.Get(ctx, name, &secret); err != nil {
			return handleError(fmt.Errorf("failed to get secret '%s': %w", name.String(), err), sourcev1.AuthenticationFailedReason)
		}

		var u *url.URL
		if u, err = url.Parse(obj.Spec.URL); err == nil {
			err = cmd.addGitEnv(dir, *u, &secret)
		}
	}
	if err != nil {
		return handleError(fmt.Errorf("failed to configure auth strategy kluctl: %w", err), sourcev1.AuthenticationFailedReason)
	}

	cmd.args = append(cmd.args, "archive")
	cmd.args = append(cmd.args, "-p", obj.Spec.URL)
	if obj.Spec.Ref != nil {
		cmd.args = append(cmd.args, "-b", *obj.Spec.Ref)
	}

	archivePath := filepath.Join(dir, "archive.tar.gz")
	metadataPath := filepath.Join(dir, "metadata.yaml")
	cmd.args = append(cmd.args, "--output-archive", archivePath)
	cmd.args = append(cmd.args, "--output-metadata", metadataPath)

	_, _, err = cmd.run(ctx)
	if err != nil {
		return handleError(fmt.Errorf("kluctl archive failed: %w", err), v1alpha1.ArchiveFailedReason)
	}

	archiveHash, err := r.calcHash(archivePath)
	if err != nil {
		return handleError(fmt.Errorf("hashing %s failed: %w", archivePath, err), v1alpha1.ArchiveFailedReason)
	}
	metadataHash, err := r.calcHash(metadataPath)
	if err != nil {
		return handleError(fmt.Errorf("hashing %s failed: %w", metadataPath, err), v1alpha1.ArchiveFailedReason)
	}
	archiveInfo.ArchiveHash = archiveHash
	archiveInfo.MetadataHash = metadataHash

	err = yaml.ReadYamlFile(metadataPath, metadata)
	if err != nil {
		return handleError(fmt.Errorf("reading %s failed: %w", metadataPath, err), v1alpha1.ArchiveFailedReason)
	}

	ctrl.LoggerFrom(ctx).V(logger.DebugLevel).Info("kluctl archive created", "url", obj.Spec.URL, "revision", archiveInfo.String())
	conditions.Delete(obj, v1alpha1.ArchiveFailedCondition)

	// Mark observations about the revision on the object
	if !obj.GetArtifact().HasRevision(archiveInfo.String()) {
		message := fmt.Sprintf("new upstream revision '%s'", archiveInfo.String())
		conditions.MarkTrue(obj, sourcev1.ArtifactOutdatedCondition, "NewRevision", message)
		conditions.MarkReconciling(obj, "NewRevision", message)
	}
	return sreconcile.ResultSuccess, nil
}

// reconcileArtifact archives a new Artifact to the Storage, if the current
// (Status) data on the object does not match the given.
//
// The inspection of the given data to the object is differed, ensuring any
// stale observations like v1beta2.ArtifactOutdatedCondition are removed.
// If the given Artifact does not differ from the
// object's current, it returns early.
// On a successful archive, the Artifact in the Status of the
// object is set, and the symlink in the Storage is updated to its path.
func (r *KluctlProjectReconciler) reconcileArtifact(ctx context.Context,
	obj *v1alpha1.KluctlProject, archiveInfo *v1alpha1.ArchiveInfo, metadata *types2.ArchiveMetadata, dir string) (sreconcile.Result, error) {
	// Create potential new artifact with current available metadata
	artifact := r.Storage.NewArtifactFor(obj.Kind, obj.GetObjectMeta(), archiveInfo.String(), fmt.Sprintf("%s.tar.gz", archiveInfo.String()))

	// Set the ArtifactInStorageCondition if there's no drift.
	defer func() {
		if obj.GetArtifact().HasRevision(artifact.Revision) {
			conditions.Delete(obj, sourcev1.ArtifactOutdatedCondition)
			conditions.MarkTrue(obj, sourcev1.ArtifactInStorageCondition, meta.SucceededReason,
				"stored artifact for revision '%s'", artifact.Revision)
		}
	}()

	// The artifact is up-to-date
	if obj.GetArtifact().HasRevision(artifact.Revision) {
		r.eventLogf(ctx, obj, events.EventTypeTrace, sourcev1.ArtifactUpToDateReason, "artifact up-to-date with remote revision: '%s'", artifact.Revision)
		return sreconcile.ResultSuccess, nil
	}

	// Ensure target path exists and is a directory
	if f, err := os.Stat(dir); err != nil {
		e := &serror.Event{
			Err:    fmt.Errorf("failed to stat target artifact path: %w", err),
			Reason: sourcev1.StatOperationFailedReason,
		}
		conditions.MarkTrue(obj, sourcev1.StorageOperationFailedCondition, e.Reason, e.Err.Error())
		return sreconcile.ResultEmpty, e
	} else if !f.IsDir() {
		e := &serror.Event{
			Err:    fmt.Errorf("invalid target path: '%s' is not a directory", dir),
			Reason: sourcev1.InvalidPathReason,
		}
		conditions.MarkTrue(obj, sourcev1.StorageOperationFailedCondition, e.Reason, e.Err.Error())
		return sreconcile.ResultEmpty, e
	}

	// Ensure artifact directory exists and acquire lock
	if err := r.Storage.MkdirAll(artifact); err != nil {
		e := &serror.Event{
			Err:    fmt.Errorf("failed to create artifact directory: %w", err),
			Reason: sourcev1.DirCreationFailedReason,
		}
		conditions.MarkTrue(obj, sourcev1.StorageOperationFailedCondition, e.Reason, e.Err.Error())
		return sreconcile.ResultEmpty, e
	}
	unlock, err := r.Storage.Lock(artifact)
	if err != nil {
		return sreconcile.ResultEmpty, &serror.Event{
			Err:    fmt.Errorf("failed to acquire lock for artifact: %w", err),
			Reason: meta.FailedReason,
		}
	}
	defer unlock()

	// Archive directory to storage
	if err := r.Storage.Archive(&artifact, dir, nil); err != nil {
		e := &serror.Event{
			Err:    fmt.Errorf("unable to archive artifact to storage: %w", err),
			Reason: sourcev1.ArchiveOperationFailedReason,
		}
		conditions.MarkTrue(obj, sourcev1.StorageOperationFailedCondition, e.Reason, e.Err.Error())
		return sreconcile.ResultEmpty, e
	}

	// Record it on the object
	obj.Status.Artifact = artifact.DeepCopy()
	obj.Status.ArchiveInfo = archiveInfo

	obj.Status.Targets = nil
	for _, t := range metadata.Targets {
		obj.Status.Targets = append(obj.Status.Targets, t.Target.Name)
	}

	obj.Status.InvolvedRepos = nil
	for u, repo := range metadata.InvolvedRepos {
		ir2 := v1alpha1.InvolvedRepo{
			URL: u,
		}

		for _, ir := range repo {
			ir2.Patterns = append(ir2.Patterns, v1alpha1.InvolvedRepoPattern{
				Pattern: ir.RefPattern,
				Refs:    ir.Refs,
			})
		}

		obj.Status.InvolvedRepos = append(obj.Status.InvolvedRepos, ir2)
	}

	// Update symlink on a "best effort" basis
	url, err := r.Storage.Symlink(artifact, "latest.tar.gz")
	if err != nil {
		r.eventLogf(ctx, obj, events.EventTypeTrace, sourcev1.SymlinkUpdateFailedReason,
			"failed to update status URL symlink: %s", err)
	}
	if url != "" {
		obj.Status.URL = url
	}
	conditions.Delete(obj, sourcev1.StorageOperationFailedCondition)
	return sreconcile.ResultSuccess, nil
}

// reconcileDelete handles the deletion of the object.
// It first garbage collects all Artifacts for the object from the Storage.
// Removing the finalizer from the object if successful.
func (r *KluctlProjectReconciler) reconcileDelete(ctx context.Context, obj *v1alpha1.KluctlProject) (sreconcile.Result, error) {
	// Garbage collect the resource's artifacts
	if err := r.garbageCollect(ctx, obj); err != nil {
		// Return the error so we retry the failed garbage collection
		return sreconcile.ResultEmpty, err
	}

	// Remove our finalizer from the list
	controllerutil.RemoveFinalizer(obj, sourcev1.SourceFinalizer)

	// Stop reconciliation as the object is being deleted
	return sreconcile.ResultEmpty, nil
}

// garbageCollect performs a garbage collection for the given object.
//
// It removes all but the current Artifact from the Storage, unless the
// deletion timestamp on the object is set. Which will result in the
// removal of all Artifacts for the objects.
func (r *KluctlProjectReconciler) garbageCollect(ctx context.Context, obj *v1alpha1.KluctlProject) error {
	if !obj.DeletionTimestamp.IsZero() {
		if deleted, err := r.Storage.RemoveAll(r.Storage.NewArtifactFor(obj.Kind, obj.GetObjectMeta(), "", "*")); err != nil {
			return &serror.Event{
				Err:    fmt.Errorf("garbage collection for deleted resource failed: %w", err),
				Reason: "GarbageCollectionFailed",
			}
		} else if deleted != "" {
			r.eventLogf(ctx, obj, events.EventTypeTrace, "GarbageCollectionSucceeded",
				"garbage collected artifacts for deleted resource")
		}
		obj.Status.Artifact = nil
		return nil
	}
	if obj.GetArtifact() != nil {
		delFiles, err := r.Storage.GarbageCollect(ctx, *obj.GetArtifact(), time.Second*5)
		if err != nil {
			return &serror.Event{
				Err:    fmt.Errorf("garbage collection of artifacts failed: %w", err),
				Reason: "GarbageCollectionFailed",
			}
		}
		if len(delFiles) > 0 {
			r.eventLogf(ctx, obj, events.EventTypeTrace, "GarbageCollectionSucceeded",
				fmt.Sprintf("garbage collected %d artifacts", len(delFiles)))
			return nil
		}
	}
	return nil
}

// eventLogf records events, and logs at the same time.
//
// This log is different from the debug log in the EventRecorder, in the sense
// that this is a simple log. While the debug log contains complete details
// about the event.
func (r *KluctlProjectReconciler) eventLogf(ctx context.Context, obj runtime.Object, eventType string, reason string, messageFmt string, args ...interface{}) {
	msg := fmt.Sprintf(messageFmt, args...)
	// Log and emit event.
	if eventType == corev1.EventTypeWarning {
		ctrl.LoggerFrom(ctx).Error(errors.New(reason), msg)
	} else {
		ctrl.LoggerFrom(ctx).Info(msg)
	}
	r.Eventf(obj, eventType, reason, msg)
}

func (r *KluctlProjectReconciler) calcHash(path string) (string, error) {
	reader, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	return r.Storage.Checksum(reader), nil
}
