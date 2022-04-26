module github.com/kluctl/flux-kluctl-controller

go 1.16

require (
	github.com/cyphar/filepath-securejoin v0.2.3
	github.com/fluxcd/pkg/apis/acl v0.0.3
	github.com/fluxcd/pkg/apis/meta v0.12.2
	github.com/fluxcd/pkg/runtime v0.14.2-0.20220414145442-395b280a83c3
	github.com/fluxcd/pkg/untar v0.1.0
	github.com/fluxcd/source-controller/api v0.24.0
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/kluctl/kluctl v1.28.2-0.20220419135732-a792dfbb27c6
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.24.0-beta.0
	k8s.io/apimachinery v0.24.0-beta.0
	k8s.io/client-go v0.24.0-beta.0
	sigs.k8s.io/controller-runtime v0.11.2
)

require (
	github.com/fluxcd/pkg/lockedfile v0.1.0
	github.com/fluxcd/source-controller v0.24.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v1.2.3
	go.uber.org/atomic v1.9.0 // indirect
)
