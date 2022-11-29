package controllers

import (
	"context"
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/e2e/test-utils"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"github.com/kluctl/kluctl/v2/pkg/utils/uo"
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

	p := test_utils.NewTestProject(t, nil)
	p.UpdateTarget("target1", nil)
	p.AddKustomizeDeployment("d1", []test_utils.KustomizeResource{
		{Name: "cm1.yaml", Content: uo.FromStringMust(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
  namespace: "{{ args.namespace }}"
data:
  k1: v1
`)},
	}, nil)

	artifactFile, artifactChecksum, err := artifactFromDir(p.LocalRepoDir())
	g.Expect(err).ToNot(HaveOccurred())

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

func TestKluctlDeploymentReconciler_Helm(t *testing.T) {
	g := NewWithT(t)
	namespace := "kluctl-helm-" + randStringRunes(5)

	p := test_utils.NewTestProject(t, nil)
	p.UpdateTarget("target1", nil)

	repoUrl := test_utils.CreateHelmRepo(t, []test_utils.RepoChart{
		{ChartName: "test-chart1", Version: "0.1.0"},
	}, "", "")
	repoUrlWithCreds := test_utils.CreateHelmRepo(t, []test_utils.RepoChart{
		{ChartName: "test-chart2", Version: "0.1.0"},
	}, "test-user", "test-password")
	ociRepoUrlWithCreds := test_utils.CreateOciRepo(t, []test_utils.RepoChart{
		{ChartName: "test-chart3", Version: "0.1.0"},
	}, "test-user", "test-password")

	p.AddHelmDeployment("d1", repoUrl, "test-chart1", "0.1.0", "test-helm-1", namespace, nil)

	err := createNamespace(namespace)
	g.Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

	artifactFile, artifactChecksum, err := artifactFromDir(p.LocalRepoDir())
	g.Expect(err).ToNot(HaveOccurred())

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

	t.Run("chart got deployed", func(t *testing.T) {
		err := k8sClient.Get(context.TODO(), client.ObjectKey{
			Name:      "test-helm-1-test-chart1",
			Namespace: namespace,
		}, cm)
		g.Expect(err).To(Succeed())
		g.Expect(cm.Data).To(HaveKeyWithValue("a", "v1"))
	})

	p.AddHelmDeployment("d2", repoUrlWithCreds, "test-chart2", "0.1.0", "test-helm-2", namespace, nil)

	artifactFile, artifactChecksum, err = artifactFromDir(p.LocalRepoDir())
	g.Expect(err).To(Succeed())
	err = applyGitRepository(repositoryName, artifactFile, "main/"+artifactChecksum)
	g.Expect(err).To(Succeed())

	t.Run("chart with credentials fails with 401", func(t *testing.T) {
		g.Eventually(func() bool {
			err = k8sClient.Get(context.TODO(), kluctlDeploymentKey, kluctlDeployment)
			g.Expect(err).To(Succeed())
			for _, c := range kluctlDeployment.Status.Conditions {
				_ = c
				if c.Type == "Ready" && c.Reason == "PrepareFailed" && strings.Contains(c.Message, "401 Unauthorized") {
					return true
				}
			}
			return false
		}, timeout, time.Second).Should(BeTrue())
	})

	credsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "helm-secrets-1",
		},
		Data: map[string][]byte{
			"url":      []byte(repoUrlWithCreds),
			"username": []byte("test-user"),
			"password": []byte("test-password"),
		},
	}
	err = k8sClient.Create(context.TODO(), credsSecret)
	g.Expect(err).To(Succeed())

	kluctlDeployment.Spec.HelmCredentials = append(kluctlDeployment.Spec.HelmCredentials, kluctlv1.HelmCredentials{SecretRef: meta.LocalObjectReference{Name: "helm-secrets-1"}})
	err = k8sClient.Update(context.TODO(), kluctlDeployment)
	g.Expect(err).To(Succeed())

	t.Run("chart with credentials succeeds", func(t *testing.T) {
		g.Eventually(func() bool {
			err := k8sClient.Get(context.TODO(), client.ObjectKey{
				Name:      "test-helm-2-test-chart2",
				Namespace: namespace,
			}, cm)
			if err != nil {
				return false
			}
			g.Expect(cm.Data).To(HaveKeyWithValue("a", "v1"))
			return true
		}, timeout, time.Second).Should(BeTrue())
	})

	p.AddHelmDeployment("d3", ociRepoUrlWithCreds, "test-chart3", "0.1.0", "test-helm-3", namespace, nil)

	artifactFile, artifactChecksum, err = artifactFromDir(p.LocalRepoDir())
	g.Expect(err).To(Succeed())
	err = applyGitRepository(repositoryName, artifactFile, "main/"+artifactChecksum)
	g.Expect(err).To(Succeed())

	t.Run("OCI chart with credentials fails with 401", func(t *testing.T) {
		g.Eventually(func() bool {
			err = k8sClient.Get(context.TODO(), kluctlDeploymentKey, kluctlDeployment)
			g.Expect(err).To(Succeed())
			for _, c := range kluctlDeployment.Status.Conditions {
				_ = c
				if c.Type == "Ready" && c.Reason == "PrepareFailed" && strings.Contains(c.Message, "401 Unauthorized") {
					return true
				}
			}
			return false
		}, timeout, time.Second).Should(BeTrue())
	})

	/*
		TODO enable this when Kluctl supports OCI authentication
		url, err := url2.Parse(ociRepoUrlWithCreds)
		g.Expect(err).To(Succeed())

		dockerJson := map[string]any{
			"auths": map[string]any{
				url.Host: map[string]any{
					"username": "test-user",
					"password": "test-password,",
					"auth":     base64.StdEncoding.EncodeToString([]byte("test-user:test-password")),
				},
			},
		}
		dockerJsonStr, err := json.Marshal(dockerJson)
		g.Expect(err).To(Succeed())

		credsSecret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "helm-secrets-2",
			},
			Data: map[string][]byte{
				"url":               []byte(ociRepoUrlWithCreds),
				".dockerconfigjson": dockerJsonStr,
			},
		}
		err = k8sClient.Create(context.TODO(), credsSecret2)
		g.Expect(err).To(Succeed())

		kluctlDeployment.Spec.HelmCredentials = append(kluctlDeployment.Spec.HelmCredentials, meta.LocalObjectReference{Name: "helm-secrets-2"})
		err = k8sClient.Update(context.TODO(), kluctlDeployment)
		g.Expect(err).To(Succeed())

		t.Run("OCI chart with credentials succeeds", func(t *testing.T) {
			g.Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), client.ObjectKey{
					Name:      "test-helm-3-test-chart3",
					Namespace: namespace,
				}, cm)
				if err != nil {
					return false
				}
				g.Expect(cm.Data).To(HaveKeyWithValue("a", "v1"))
				return true
			}, timeout, time.Second).Should(BeTrue())
		})*/
}
