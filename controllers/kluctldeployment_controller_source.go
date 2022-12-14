package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/acl"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/git/auth"
	"github.com/kluctl/kluctl/v2/pkg/git/messages"
	"github.com/kluctl/kluctl/v2/pkg/repocache"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *KluctlDeploymentReconciler) getProjectSource(ctx context.Context, obj *kluctlv1.KluctlDeployment, noCrossNamespaceRefs bool) (*kluctlv1.ProjectSource, error) {
	if obj.Spec.Source != nil && obj.Spec.SourceRef != nil {
		return nil, fmt.Errorf("sourceRef and source can't be specified at the same time")
	}
	if obj.Spec.Source == nil && obj.Spec.SourceRef == nil {
		return nil, fmt.Errorf("source not specified")
	}
	if obj.Spec.Path != "" && obj.Spec.Source != nil {
		return nil, fmt.Errorf("path and source can't be specified at the same time")
	}

	var sourceSpec kluctlv1.ProjectSource

	if obj.Spec.SourceRef != nil {
		source, err := r.getDeprecatedSource(ctx, *obj.Spec.SourceRef, obj.GetNamespace(), noCrossNamespaceRefs)
		if err != nil {
			return nil, err
		}
		if source.Spec.Reference.Commit != "" || source.Spec.Reference.SemVer != "" {
			return nil, fmt.Errorf("commit and semVer are not supported as git ref")
		}
		sourceSpec = kluctlv1.ProjectSource{
			URL: source.Spec.URL,
		}
		if source.Spec.Reference != nil {
			sourceSpec.Ref = &kluctlv1.GitRef{
				Branch: source.Spec.Reference.Branch,
				Tag:    source.Spec.Reference.Tag,
				//Commit: source.Spec.Reference.Commit,
			}
		}
		if source.Spec.SecretRef != nil {
			sourceSpec.SecretRef = &meta.LocalObjectReference{
				Name: source.Spec.SecretRef.Name,
			}
		}
	} else {
		sourceSpec = *obj.Spec.Source
	}
	if obj.Spec.Path != "" {
		sourceSpec.Path = obj.Spec.Path
	}

	return &sourceSpec, nil
}

func (r *KluctlDeploymentReconciler) getDeprecatedSource(ctx context.Context, ref meta.NamespacedObjectKindReference, holderNs string, noCrossNamespaceRefs bool) (*sourcev1.GitRepository, error) {
	var source *sourcev1.GitRepository
	sourceNamespace := holderNs
	if ref.Namespace != "" {
		sourceNamespace = ref.Namespace
	}
	namespacedName := types.NamespacedName{
		Namespace: sourceNamespace,
		Name:      ref.Name,
	}

	if noCrossNamespaceRefs && sourceNamespace != holderNs {
		return source, acl.AccessDeniedError(
			fmt.Sprintf("can't access '%s/%s', cross-namespace references have been blocked",
				ref.Kind, namespacedName))
	}

	switch ref.Kind {
	case sourcev1.GitRepositoryKind:
		var repository sourcev1.GitRepository
		err := r.Get(ctx, namespacedName, &repository)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return source, err
			}
			return source, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		source = &repository
	default:
		return source, fmt.Errorf("source `%s` kind '%s' not supported",
			ref.Name, ref.Kind)
	}
	return source, nil
}

func (r *KluctlDeploymentReconciler) getGitSecret(ctx context.Context, source *kluctlv1.ProjectSource, objNs string) (*corev1.Secret, error) {
	if source == nil || source.SecretRef == nil {
		return nil, nil
	}

	// Attempt to retrieve secret
	name := types.NamespacedName{
		Namespace: objNs,
		Name:      source.SecretRef.Name,
	}
	var secret corev1.Secret
	if err := r.Get(ctx, name, &secret); err != nil {
		return nil, fmt.Errorf("failed to get secret '%s': %w", name.String(), err)
	}
	return &secret, nil
}

func (r *KluctlDeploymentReconciler) buildGitAuth(ctx context.Context, gitSecret *corev1.Secret) (*auth.GitAuthProviders, error) {
	log := ctrl.LoggerFrom(ctx)
	ga := auth.NewDefaultAuthProviders("KLUCTL_GIT", &messages.MessageCallbacks{
		WarningFn: func(s string) {
			log.Info(s)
		},
		TraceFn: func(s string) {
			log.V(1).Info(s)
		},
	})

	if gitSecret == nil {
		return ga, nil
	}

	e := auth.AuthEntry{
		Host:     "*",
		Username: "*",
	}

	if x, ok := gitSecret.Data["username"]; ok {
		e.Username = string(x)
	}
	if x, ok := gitSecret.Data["password"]; ok {
		e.Password = string(x)
	}
	if x, ok := gitSecret.Data["caFile"]; ok {
		e.CABundle = x
	}
	if x, ok := gitSecret.Data["known_hosts"]; ok {
		e.KnownHosts = x
	}
	if x, ok := gitSecret.Data["identity"]; ok {
		e.SshKey = x
	}

	var la auth.ListAuthProvider
	la.AddEntry(e)
	ga.RegisterAuthProvider(&la, false)
	return ga, nil
}

func (r *KluctlDeploymentReconciler) buildRepoCache(ctx context.Context, secret *corev1.Secret) (*repocache.GitRepoCache, error) {
	// make sure we use a unique repo cache per set of credentials
	h := sha256.New()
	if secret == nil {
		h.Write([]byte("no-secret"))
	} else {
		m := json.NewEncoder(h)
		err := m.Encode(secret.Data)
		if err != nil {
			return nil, err
		}
	}
	h2 := hex.EncodeToString(h.Sum(nil))

	tmpBaseDir := filepath.Join(os.TempDir(), "kluctl-controller-repo-cache", h2)
	err := os.MkdirAll(tmpBaseDir, 0o700)
	if err != nil {
		return nil, err
	}

	ctx = utils.WithTmpBaseDir(ctx, tmpBaseDir)

	ga, err := r.buildGitAuth(ctx, secret)
	if err != nil {
		return nil, err
	}

	rc := repocache.NewGitRepoCache(ctx, r.SshPool, ga, nil, 0)
	return rc, nil
}
