package controllers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/acl"
	"github.com/fluxcd/pkg/untar"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/kluctl/kluctl/v2/pkg/git/auth"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func (r *KluctlDeploymentReconciler) getSource(ctx context.Context, ref meta.NamespacedObjectKindReference, holderNs string, noCrossNamespaceRefs bool) (sourcev1.Source, error) {
	var source sourcev1.Source
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

func (r *KluctlDeploymentReconciler) download(source sourcev1.Source, tmpDir string) error {
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

func (r *KluctlDeploymentReconciler) verifyArtifact(artifact *sourcev1.Artifact, buf *bytes.Buffer, reader io.Reader) error {
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

func (r *KluctlDeploymentReconciler) getGitSecret(ctx context.Context, source sourcev1.Source, objNs string) (*corev1.Secret, error) {
	if gitRepo, ok := source.(*sourcev1.GitRepository); ok {
		if gitRepo.Spec.SecretRef == nil {
			return nil, nil
		}
		// Attempt to retrieve secret
		name := types.NamespacedName{
			Namespace: objNs,
			Name:      gitRepo.Spec.SecretRef.Name,
		}
		var secret corev1.Secret
		if err := r.Get(ctx, name, &secret); err != nil {
			return nil, fmt.Errorf("failed to get secret '%s': %w", name.String(), err)
		}
		return &secret, nil
	}
	return nil, nil
}

func (r *KluctlDeploymentReconciler) buildGitAuth(ctx context.Context, source sourcev1.Source, objNs string) (*auth.GitAuthProviders, error) {
	ga := auth.NewDefaultAuthProviders()

	gitSecret, err := r.getGitSecret(ctx, source, objNs)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

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
