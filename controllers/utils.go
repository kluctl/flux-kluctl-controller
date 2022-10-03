package controllers

import (
	"github.com/fluxcd/pkg/apis/meta"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setReadiness(obj *kluctlv1.KluctlDeployment, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  status,
		Reason:  reason,
		Message: trimString(message, kluctlv1.MaxConditionMessageLength),
	}

	c := obj.GetConditions()
	apimeta.SetStatusCondition(&c, newCondition)
	obj.SetConditions(c)

	obj.Status.ObservedGeneration = obj.GetGeneration()
}

func setReadinessWithRevision(obj *kluctlv1.KluctlDeployment, status metav1.ConditionStatus, reason, message string, revision string) {
	setReadiness(obj, status, reason, message)
	obj.Status.LastAttemptedRevision = revision
}

func trimString(str string, limit int) string {
	if len(str) <= limit {
		return str
	}

	return str[0:limit] + "..."
}
