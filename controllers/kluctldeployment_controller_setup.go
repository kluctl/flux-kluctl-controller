package controllers

import (
	"fmt"
	"github.com/fluxcd/pkg/runtime/predicates"
	"github.com/hashicorp/go-retryablehttp"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// SetupWithManager sets up the controller with the Manager.
func (r *KluctlDeploymentReconciler) SetupWithManager(mgr ctrl.Manager, opts KluctlDeploymentReconcilerOpts) error {
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
			predicate.Or(predicate.GenerationChangedPredicate{}, predicates.ReconcileRequestedPredicate{}, DeployRequestedPredicate{}),
		)).
		WithOptions(controller.Options{MaxConcurrentReconciles: opts.MaxConcurrentReconciles}).
		Complete(r)
}
