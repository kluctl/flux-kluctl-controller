package controllers

import (
	"context"
	"fmt"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/acl"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSource(ctx context.Context, client client.Client, ref meta.NamespacedObjectKindReference, holderNs string, noCrossNamespaceRefs bool) (sourcev1.Source, error) {
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
		err := client.Get(ctx, namespacedName, &repository)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return source, err
			}
			return source, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		source = &repository
	case sourcev1.BucketKind:
		var bucket sourcev1.Bucket
		err := client.Get(ctx, namespacedName, &bucket)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return source, err
			}
			return source, fmt.Errorf("unable to get source '%s': %w", namespacedName, err)
		}
		source = &bucket
	default:
		return source, fmt.Errorf("source `%s` kind '%s' not supported",
			ref.Name, ref.Kind)
	}
	return source, nil
}
