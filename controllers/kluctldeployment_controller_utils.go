package controllers

import (
	"context"
	fluxv1beta1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *KluctlDeploymentReconciler) event(ctx context.Context, obj *kluctlv1.KluctlDeployment, revision, severity, msg string, metadata map[string]string) {
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
	if severity == fluxv1beta1.EventSeverityError {
		eventtype = "Warning"
	}

	r.EventRecorder.AnnotatedEventf(obj, metadata, eventtype, reason, msg)
}

func (r *KluctlDeploymentReconciler) recordReadiness(ctx context.Context, obj *kluctlv1.KluctlDeployment) {
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

func (r *KluctlDeploymentReconciler) recordSuspension(ctx context.Context, obj *kluctlv1.KluctlDeployment) {
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
		r.MetricsRecorder.RecordSuspend(*objRef, obj.Spec.Suspend)
	}
}
