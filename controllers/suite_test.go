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
	"fmt"
	test_utils "github.com/kluctl/kluctl/v2/e2e/test-utils"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func getHeadRevision(t *testing.T, p *test_utils.TestProject) string {
	r := p.GetGitRepo()
	h, err := r.Head()
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s/%s", h.Name().String(), h.Hash().String())
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
