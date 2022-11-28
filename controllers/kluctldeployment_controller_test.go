package controllers

import (
	"context"
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"
	"time"
)

func TestKluctlDeploymentReconciler_FieldManager(t *testing.T) {
	g := NewWithT(t)
	namespace := "kluctl-fieldmanager-" + randStringRunes(5)

	err := createNamespace(namespace)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

	artifactFile, err := artifactFromDir("testdata/targets")
	g.Expect(err).ToNot(HaveOccurred())

	artifactChecksum := strings.TrimSuffix(artifactFile, ".tar.gz")

	repositoryName := types.NamespacedName{
		Name:      randStringRunes(5),
		Namespace: namespace,
	}

	err = applyGitRepository(repositoryName, artifactFile, "main/"+artifactChecksum)
	g.Expect(err).NotTo(HaveOccurred())

	kluctlDeploymentKey := types.NamespacedName{
		Name:      "kluctl-fieldmanager-" + randStringRunes(5),
		Namespace: namespace,
	}
	kluctlDeployment := &kluctlv1.KluctlDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kluctlDeploymentKey.Name,
			Namespace: kluctlDeploymentKey.Namespace,
		},
		Spec: kluctlv1.KluctlDeploymentSpec{
			Interval: metav1.Duration{Duration: reconciliationInterval},
			Timeout:  &metav1.Duration{Duration: timeout},
			Target:   utils.StrPtr("target1"),
			Args: runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{"namespace": "%s"}`, namespace)),
			},
			SourceRef: meta.NamespacedObjectKindReference{
				Name:      repositoryName.Name,
				Namespace: repositoryName.Namespace,
				Kind:      sourcev1.GitRepositoryKind,
			},
		},
	}

	g.Expect(k8sClient.Create(context.TODO(), kluctlDeployment)).To(Succeed())

	g.Eventually(func() bool {
		var obj kluctlv1.KluctlDeployment
		_ = k8sClient.Get(context.Background(), client.ObjectKeyFromObject(kluctlDeployment), &obj)
		if obj.Status.LastDeployResult == nil {
			return false
		}
		return obj.Status.LastDeployResult.Revision == "main/"+artifactChecksum
	}, timeout, time.Second).Should(BeTrue())

	cm := &corev1.ConfigMap{}

	t.Run("cm1 is deployed", func(t *testing.T) {
		err := k8sClient.Get(context.TODO(), client.ObjectKey{
			Name:      "cm1",
			Namespace: namespace,
		}, cm)
		g.Expect(err).To(Succeed())
		g.Expect(cm.Data).To(HaveKeyWithValue("k1", "v1"))
	})

	t.Run("cm1 is modified and restored", func(t *testing.T) {
		cm.Data["k1"] = "v2"
		err := k8sClient.Update(context.TODO(), cm, client.FieldOwner("kubectl"))
		g.Expect(err).To(Succeed())

		g.Eventually(func() bool {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      "cm1",
				Namespace: namespace,
			}, cm)
			g.Expect(err).To(Succeed())
			return cm.Data["k1"] == "v1"
		}, timeout, time.Second).Should(BeTrue())
	})

	t.Run("cm1 gets a key added which is not modified by the controller", func(t *testing.T) {
		cm.Data["k1"] = "v2"
		cm.Data["k2"] = "v2"
		err := k8sClient.Update(context.TODO(), cm, client.FieldOwner("kubectl"))
		g.Expect(err).To(Succeed())

		g.Eventually(func() bool {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      "cm1",
				Namespace: namespace,
			}, cm)
			g.Expect(err).To(Succeed())
			return cm.Data["k1"] == "v1"
		}, timeout, time.Second).Should(BeTrue())

		g.Expect(cm.Data).To(HaveKeyWithValue("k2", "v2"))
	})

	t.Run("cm1 gets modified with another field manager", func(t *testing.T) {
		patch := client.MergeFrom(cm.DeepCopy())
		cm.Data["k1"] = "v2"

		err := k8sClient.Patch(context.TODO(), cm, patch, client.FieldOwner("test-field-manager"))
		g.Expect(err).To(Succeed())

		for i := 0; i < 2; i++ {
			waitForReconcile(t, kluctlDeploymentKey)
		}

		err = k8sClient.Get(context.TODO(), client.ObjectKey{
			Name:      "cm1",
			Namespace: namespace,
		}, cm)
		g.Expect(err).To(Succeed())
		g.Expect(cm.Data).To(HaveKeyWithValue("k1", "v2"))
	})

	err = k8sClient.Get(context.TODO(), kluctlDeploymentKey, kluctlDeployment)
	g.Expect(err).To(Succeed())

	kluctlDeployment.Spec.ForceApply = true
	err = k8sClient.Update(context.TODO(), kluctlDeployment)
	g.Expect(err).To(Succeed())

	t.Run("forceApply is true and cm1 gets restored even with another field manager", func(t *testing.T) {
		patch := client.MergeFrom(cm.DeepCopy())
		cm.Data["k1"] = "v2"

		err := k8sClient.Patch(context.TODO(), cm, patch, client.FieldOwner("test-field-manager"))
		g.Expect(err).To(Succeed())

		g.Eventually(func() bool {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      "cm1",
				Namespace: namespace,
			}, cm)
			g.Expect(err).To(Succeed())
			return cm.Data["k1"] == "v1"
		}, timeout, time.Second).Should(BeTrue())
	})
}
