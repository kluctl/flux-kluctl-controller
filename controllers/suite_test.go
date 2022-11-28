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
	"crypto/sha256"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/controller"
	"github.com/fluxcd/pkg/runtime/testenv"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	controllerLog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluxcd/pkg/testserver"

	kluctliov1alpha1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	timeout                = time.Second * 30
	interval               = time.Second * 1
	reconciliationInterval = time.Second * 5
)

var (
	reconciler   *KluctlDeploymentReconciler
	k8sClient    client.Client
	testEnv      *testenv.Environment
	testServer   *testserver.ArtifactServer
	testMetricsH controller.Metrics
	ctx          = ctrl.SetupSignalHandler()
	kubeConfig   []byte
	debugMode    = os.Getenv("DEBUG_TEST") != ""
)

func runInContext(registerControllers func(*testenv.Environment), run func() error, crdPath string) error {
	var err error
	utilruntime.Must(sourcev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(kluctliov1alpha1.AddToScheme(scheme.Scheme))

	if debugMode {
		controllerLog.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(false)))
	}

	testEnv = testenv.New(testenv.WithCRDPath(crdPath))

	testServer, err = testserver.NewTempArtifactServer()
	if err != nil {
		panic(fmt.Sprintf("Failed to create a temporary storage server: %v", err))
	}
	fmt.Println("Starting the test storage server")
	testServer.Start()

	registerControllers(testEnv)

	go func() {
		fmt.Println("Starting the test environment")
		if err := testEnv.Start(ctx); err != nil {
			panic(fmt.Sprintf("Failed to start the test environment manager: %v", err))
		}
	}()
	<-testEnv.Manager.Elected()

	user, err := testEnv.AddUser(envtest.User{
		Name:   "testenv-admin",
		Groups: []string{"system:masters"},
	}, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to create testenv-admin user: %v", err))
	}

	kubeConfig, err = user.KubeConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to create the testenv-admin user kubeconfig: %v", err))
	}

	// Client with caching disabled.
	k8sClient, err = client.New(testEnv.Config, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		panic(fmt.Sprintf("Failed to create k8s client: %v", err))
	}

	runErr := run()

	if debugMode {
		events := &corev1.EventList{}
		_ = k8sClient.List(ctx, events)
		for _, event := range events.Items {
			fmt.Printf("%s %s \n%s\n",
				event.InvolvedObject.Name, event.GetAnnotations()["kustomize.toolkit.fluxcd.io/revision"],
				event.Message)
		}
	}

	fmt.Println("Stopping the test environment")
	if err := testEnv.Stop(); err != nil {
		panic(fmt.Sprintf("Failed to stop the test environment: %v", err))
	}

	fmt.Println("Stopping the file server")
	testServer.Stop()
	if err := os.RemoveAll(testServer.Root()); err != nil {
		panic(fmt.Sprintf("Failed to remove storage server dir: %v", err))
	}

	return runErr
}

func TestMain(m *testing.M) {
	code := 0

	runInContext(func(testEnv *testenv.Environment) {
		controllerName := "flux-kluctl-controller"
		testMetricsH = controller.MustMakeMetrics(testEnv)
		reconciler = &KluctlDeploymentReconciler{
			ControllerName:  controllerName,
			RestConfig:      testEnv.Config,
			Client:          testEnv,
			EventRecorder:   testEnv.GetEventRecorderFor(controllerName),
			MetricsRecorder: testMetricsH.MetricsRecorder,
		}
		if err := (reconciler).SetupWithManager(testEnv, KluctlDeploymentReconcilerOpts{
			MaxConcurrentReconciles: 4,
		}); err != nil {
			panic(fmt.Sprintf("Failed to start KustomizationReconciler: %v", err))
		}
	}, func() error {
		code = m.Run()
		return nil
	}, filepath.Join("..", "config", "crd", "bases"))

	os.Exit(code)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createNamespace(name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	return k8sClient.Create(context.Background(), namespace)
}

func artifactFromDir(dir string) (string, error) {
	var files []testserver.File
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		f, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, testserver.File{
			Name: relPath,
			Body: string(f),
		})
		return nil
	})
	if err != nil {
		return "", err
	}
	return testServer.ArtifactFromFiles(files)
}

func applyGitRepository(objKey client.ObjectKey, artifactName string, revision string) error {
	repo := &sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.GitRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      objKey.Name,
			Namespace: objKey.Namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL:      "https://github.com/test/repository",
			Interval: metav1.Duration{Duration: time.Minute},
		},
	}

	b, _ := os.ReadFile(filepath.Join(testServer.Root(), artifactName))
	checksum := fmt.Sprintf("%x", sha256.Sum256(b))

	url := fmt.Sprintf("%s/%s", testServer.URL(), artifactName)

	status := sourcev1.GitRepositoryStatus{
		Conditions: []metav1.Condition{
			{
				Type:               meta.ReadyCondition,
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             sourcev1.GitOperationSucceedReason,
			},
		},
		Artifact: &sourcev1.Artifact{
			Path:           url,
			URL:            url,
			Revision:       revision,
			Checksum:       checksum,
			LastUpdateTime: metav1.Now(),
		},
	}

	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("flux-kluctl-controller"),
	}

	if err := k8sClient.Patch(context.Background(), repo, client.Apply, opt...); err != nil {
		return err
	}

	repo.ManagedFields = nil
	repo.Status = status
	if err := k8sClient.Status().Patch(context.Background(), repo, client.Apply, opt...); err != nil {
		return err
	}
	return nil
}

func waitForReconcile(t *testing.T, key types.NamespacedName) {
	g := NewWithT(t)

	var kd kluctliov1alpha1.KluctlDeployment
	err := k8sClient.Get(context.TODO(), key, &kd)
	g.Expect(err).To(Succeed())

	lastReconcileTime := kd.Status.LastDeployResult.AttemptedAt
	g.Eventually(func() bool {
		err = k8sClient.Get(context.TODO(), key, &kd)
		g.Expect(err).To(Succeed())
		return kd.Status.LastDeployResult.AttemptedAt != lastReconcileTime
	}, timeout, time.Second).Should(BeTrue())
}
