package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/deployment"
	"github.com/kluctl/kluctl/v2/pkg/deployment/commands"
	utils2 "github.com/kluctl/kluctl/v2/pkg/deployment/utils"
	"github.com/kluctl/kluctl/v2/pkg/git/auth"
	"github.com/kluctl/kluctl/v2/pkg/jinja2"
	k8s2 "github.com/kluctl/kluctl/v2/pkg/k8s"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	"github.com/kluctl/kluctl/v2/pkg/registries"
	types2 "github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
)

type preparedProject struct {
	r      *KluctlDeploymentReconciler
	d      kluctlv1.KluctlDeployment
	source sourcev1.Source

	tmpDir     string
	repoDir    string
	projectDir string

	gitRepo   *sourcev1.GitRepository
	gitSecret *corev1.Secret

	metadata   types2.ProjectMetadata
	target     *types2.Target
	targetHash string
}

func prepareProject(ctx context.Context,
	r *KluctlDeploymentReconciler,
	kluctlDeployment kluctlv1.KluctlDeployment,
	source sourcev1.Source) (*preparedProject, error) {

	pp := &preparedProject{
		r:      r,
		d:      kluctlDeployment,
		source: source,
	}

	// create tmp dir
	tmpDir, err := os.MkdirTemp("", kluctlDeployment.Name)
	if err != nil {
		return pp, fmt.Errorf("failed to create temp dir for kluctl project: %w", err)
	}
	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		_ = os.RemoveAll(tmpDir)
	}()

	pp.tmpDir = tmpDir

	// download artifact and extract files
	err = pp.r.download(source, tmpDir)
	if err != nil {
		return pp, fmt.Errorf("failed download of artifact: %w", err)
	}

	pp.repoDir = filepath.Join(tmpDir, "source")

	// check kluctl project path exists
	pp.projectDir, err = securejoin.SecureJoin(pp.repoDir, kluctlDeployment.Spec.Path)
	if err != nil {
		return pp, err
	}
	if _, err := os.Stat(pp.projectDir); err != nil {
		return pp, fmt.Errorf("kluctlDeployment path not found: %w", err)
	}

	pp.gitSecret, err = pp.getGitSecret(ctx)
	if err != nil {
		return pp, fmt.Errorf("failed to get git secret: %w", err)
	}

	err = pp.kluctlArchive(ctx)
	if err != nil {
		return pp, err
	}

	err = yaml.ReadYamlFile(filepath.Join(pp.tmpDir, "metadata.yaml"), &pp.metadata)
	if err != nil {
		return pp, fmt.Errorf("failed to read metadata.yaml: %w", err)
	}

	for _, t := range pp.metadata.Targets {
		if t.Target.Name == kluctlDeployment.Spec.Target {
			pp.target = t.Target
			break
		}
	}
	if pp.target == nil {
		return pp, fmt.Errorf("target %s not found in kluctl project", kluctlDeployment.Spec.Target)
	}

	archiveHash, err := calcFileHash(filepath.Join(pp.tmpDir, "archive.tar.gz"))
	if err != nil {
		return pp, err
	}
	pp.targetHash = calcTargetHash(archiveHash, pp.target)

	cleanup = false
	return pp, nil
}

func (pp *preparedProject) cleanup() {
	_ = os.RemoveAll(pp.tmpDir)
}

func (pp *preparedProject) restConfigToKubeconfig(restConfig *rest.Config) *api.Config {
	kubeConfig := api.NewConfig()
	cluster := api.NewCluster()
	cluster.Server = restConfig.Host
	cluster.CertificateAuthority = restConfig.TLSClientConfig.CAFile
	cluster.CertificateAuthorityData = restConfig.TLSClientConfig.CAData
	cluster.InsecureSkipTLSVerify = restConfig.TLSClientConfig.Insecure
	kubeConfig.Clusters["default"] = cluster

	user := api.NewAuthInfo()
	user.ClientKey = restConfig.KeyFile
	user.ClientKeyData = restConfig.KeyData
	user.ClientCertificate = restConfig.CertFile
	user.ClientCertificateData = restConfig.CertData
	user.TokenFile = restConfig.BearerTokenFile
	user.Token = restConfig.BearerToken
	user.Impersonate = restConfig.Impersonate.UserName
	user.ImpersonateUID = restConfig.Impersonate.UID
	user.ImpersonateUserExtra = restConfig.Impersonate.Extra
	user.ImpersonateGroups = restConfig.Impersonate.Groups
	user.Username = restConfig.Username
	user.Password = restConfig.Password
	user.AuthProvider = restConfig.AuthProvider
	kubeConfig.AuthInfos["default"] = user

	kctx := api.NewContext()
	kctx.Cluster = "default"
	kctx.AuthInfo = "default"
	kubeConfig.Contexts["default"] = kctx

	return kubeConfig
}

func (pp *preparedProject) getKubeconfigFromSecret(ctx context.Context) ([]byte, error) {
	secretName := types.NamespacedName{
		Namespace: pp.d.GetNamespace(),
		Name:      pp.d.Spec.KubeConfig.SecretRef.Name,
	}

	var secret corev1.Secret
	if err := pp.r.Get(ctx, secretName, &secret); err != nil {
		return nil, fmt.Errorf("unable to read KubeConfig secret '%s' error: %w", secretName.String(), err)
	}

	var kubeConfig []byte
	switch {
	case pp.d.Spec.KubeConfig.SecretRef.Key != "":
		key := pp.d.Spec.KubeConfig.SecretRef.Key
		kubeConfig = secret.Data[key]
		if kubeConfig == nil {
			return nil, fmt.Errorf("KubeConfig secret '%s' does not contain a '%s' key with a kubeconfig", secretName, key)
		}
	case secret.Data["value"] != nil:
		kubeConfig = secret.Data["value"]
	case secret.Data["value.yaml"] != nil:
		kubeConfig = secret.Data["value.yaml"]
	default:
		// User did not specify a key, and the 'value' key was not defined.
		return nil, fmt.Errorf("KubeConfig secret '%s' does not contain a 'value' key with a kubeconfig", secretName)
	}

	return kubeConfig, nil
}

func (pp *preparedProject) setImpersonationConfig(restConfig *rest.Config) {
	name := pp.r.DefaultServiceAccount
	if sa := pp.d.Spec.ServiceAccountName; sa != "" {
		name = sa
	}
	if name != "" {
		username := fmt.Sprintf("system:serviceaccount:%s:%s", pp.d.GetNamespace(), name)
		restConfig.Impersonate = rest.ImpersonationConfig{UserName: username}
	}
}

func (pp *preparedProject) buildRestConfig(ctx context.Context) (*rest.Config, error) {
	var err error
	var restConfig *rest.Config

	if pp.d.Spec.KubeConfig != nil {
		kubeConfig, err := pp.getKubeconfigFromSecret(ctx)
		if err != nil {
			return nil, err
		}
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeConfig)
		if err != nil {
			return nil, err
		}
	} else {
		restConfig, err = config.GetConfig()
		if err != nil {
			return nil, err
		}
	}

	pp.setImpersonationConfig(restConfig)

	return restConfig, nil
}

func (pp *preparedProject) buildKubeconfig(ctx context.Context) (*api.Config, error) {
	restConfig, err := pp.buildRestConfig(ctx)
	if err != nil {
		return nil, err
	}

	kubeConfig := pp.restConfigToKubeconfig(restConfig)

	err = pp.renameContexts(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

func (pp *preparedProject) renameContexts(cfg *api.Config) error {
	for _, r := range pp.d.Spec.RenameContexts {
		ctx, ok := cfg.Contexts[r.OldContext]
		if !ok {
			return fmt.Errorf("failed to rename context %s -> %s. Old context not found in kubeconfig", r.OldContext, r.NewContext)
		}
		if _, ok := cfg.Contexts[r.NewContext]; ok {
			return fmt.Errorf("failed to rename context %s -> %s. New context already present in kubeconfig", r.OldContext, r.NewContext)
		}
		cfg.Contexts[r.NewContext] = ctx
		delete(cfg.Contexts, r.OldContext)

		if cfg.CurrentContext == r.OldContext {
			cfg.CurrentContext = r.NewContext
		}
	}
	return nil
}

func (pp *preparedProject) getGitSecret(ctx context.Context) (*corev1.Secret, error) {
	if gitRepo, ok := pp.source.(*sourcev1.GitRepository); ok {
		if gitRepo.Spec.SecretRef == nil {
			return nil, nil
		}
		// Attempt to retrieve secret
		name := types.NamespacedName{
			Namespace: pp.d.GetNamespace(),
			Name:      gitRepo.Spec.SecretRef.Name,
		}
		var secret corev1.Secret
		if err := pp.r.Client.Get(ctx, name, &secret); err != nil {
			return nil, fmt.Errorf("failed to get secret '%s': %w", name.String(), err)
		}
		return &secret, nil
	}
	return nil, nil
}

func (pp *preparedProject) buildGitAuth() (*auth.GitAuthProviders, error) {
	ga := auth.NewDefaultAuthProviders()
	if pp.gitSecret == nil {
		return ga, nil
	}

	e := auth.AuthEntry{
		Host:     "*",
		Username: "*",
	}

	if x, ok := pp.gitSecret.Data["username"]; ok {
		e.Username = string(x)
	}
	if x, ok := pp.gitSecret.Data["password"]; ok {
		e.Password = string(x)
	}
	if x, ok := pp.gitSecret.Data["caFile"]; ok {
		e.CABundle = x
	}
	if x, ok := pp.gitSecret.Data["known_hosts"]; ok {
		e.KnownHosts = x
	}
	if x, ok := pp.gitSecret.Data["identity"]; ok {
		e.SshKey = x
	}

	var la auth.ListAuthProvider
	la.AddEntry(e)
	ga.RegisterAuthProvider(&la, false)
	return ga, nil
}

func (pp *preparedProject) getRegistrySecrets(ctx context.Context) ([]*corev1.Secret, error) {
	var ret []*corev1.Secret
	for _, ref := range pp.d.Spec.RegistrySecrets {
		name := types.NamespacedName{
			Namespace: pp.d.GetNamespace(),
			Name:      ref.Name,
		}
		var secret corev1.Secret
		if err := pp.r.Client.Get(ctx, name, &secret); err != nil {
			return nil, fmt.Errorf("failed to get secret '%s': %w", name.String(), err)
		}
		ret = append(ret, &secret)
	}
	return ret, nil
}

func (pp *preparedProject) buildRegistryHelper(ctx context.Context) (*registries.RegistryHelper, error) {
	secrets, err := pp.getRegistrySecrets(ctx)
	if err != nil {
		return nil, err
	}

	rh := registries.NewRegistryHelper(ctx)
	err = rh.ParseAuthEntriesFromEnv()
	if err != nil {
		return nil, err
	}

	for _, s := range secrets {
		caFile := s.Data["caFile"]
		insecure := false
		if x, ok := s.Data["insecure"]; ok {
			insecure, err = strconv.ParseBool(string(x))
			if err != nil {
				return nil, fmt.Errorf("failed parsing insecure flag from secret %s: %w", s.Name, err)
			}
		}

		if dockerConfig, ok := s.Data[".dockerconfigjson"]; ok {
			maxFields := 1
			if _, ok := s.Data["caFile"]; ok {
				maxFields++
			}
			if _, ok := s.Data["insecure"]; ok {
				maxFields++
			}
			if len(s.Data) != maxFields {
				return nil, fmt.Errorf("when using .dockerconfigjson in registry secret, only caFile and insecure fields are allowed additionally")
			}

			c := configfile.New(".dockerconfigjson")
			err = c.LoadFromReader(bytes.NewReader(dockerConfig))
			if err != nil {
				return nil, fmt.Errorf("failed to parse .dockerconfigjson from secret %s: %w", s.Name, err)
			}
			for registry, ac := range c.GetAuthConfigs() {
				var e registries.AuthEntry
				e.Registry = registry
				e.Username = ac.Username
				e.Password = ac.Password
				e.Auth = ac.Auth
				e.CABundle = caFile
				e.Insecure = insecure

				rh.AddAuthEntry(e)
			}
		} else {
			var e registries.AuthEntry
			e.Registry = string(s.Data["registry"])
			e.Username = string(s.Data["username"])
			e.Password = string(s.Data["password"])
			e.Auth = string(s.Data["auth"])
			e.CABundle = caFile
			e.Insecure = insecure

			if e.Registry == "" || (e.Username == "" && e.Auth == "") {
				return nil, fmt.Errorf("registry secret is incomplete")
			}
			rh.AddAuthEntry(e)
		}
	}
	return rh, nil
}

func (pp *preparedProject) buildImages(ctx context.Context) (*deployment.Images, error) {
	rh, err := pp.buildRegistryHelper(ctx)
	if err != nil {
		return nil, err
	}
	images, err := deployment.NewImages(rh, pp.d.Spec.UpdateImages, false)
	if err != nil {
		return nil, err
	}
	for _, fi := range kluctlv1.ConvertFixedImagesToKluctl(pp.d.Spec.Images) {
		images.AddFixedImage(fi)
	}
	return images, nil
}

func (pp *preparedProject) buildInclusion() *utils.Inclusion {
	inc := utils.NewInclusion()
	for _, x := range pp.d.Spec.IncludeTags {
		inc.AddInclude("tag", x)
	}
	for _, x := range pp.d.Spec.ExcludeTags {
		inc.AddExclude("tag", x)
	}
	for _, x := range pp.d.Spec.IncludeDeploymentDirs {
		inc.AddInclude("deploymentItemDir", x)
	}
	for _, x := range pp.d.Spec.ExcludeDeploymentDirs {
		inc.AddExclude("deploymentItemDir", x)
	}
	return inc
}

func (pp *preparedProject) withKluctlProject(ctx context.Context, fromArchive bool, cb func(p *kluctl_project.LoadedKluctlProject) error) error {
	j2, err := jinja2.NewJinja2()
	if err != nil {
		return err
	}
	defer j2.Close()

	ga, err := pp.buildGitAuth()
	if err != nil {
		return err
	}

	loadArgs := kluctl_project.LoadKluctlProjectArgs{
		RepoRoot:         pp.repoDir,
		ProjectDir:       pp.projectDir,
		GitAuthProviders: ga,
	}
	if fromArchive {
		loadArgs.FromArchive = filepath.Join(pp.tmpDir, "archive.tar.gz")
		loadArgs.FromArchiveMetadata = filepath.Join(pp.tmpDir, "metadata.yaml")
	}

	p, err := kluctl_project.LoadKluctlProject(ctx, loadArgs, filepath.Join(pp.tmpDir, "project"), j2)
	if err != nil {
		return err
	}

	return cb(p)
}

func (pp *preparedProject) withKluctlProjectTarget(ctx context.Context, fromArchive bool, cb func(targetContext *kluctl_project.TargetContext) error) error {
	return pp.withKluctlProject(ctx, fromArchive, func(p *kluctl_project.LoadedKluctlProject) error {
		renderOutputDir, err := os.MkdirTemp(pp.tmpDir, "render-")
		if err != nil {
			return err
		}
		images, err := pp.buildImages(ctx)
		if err != nil {
			return err
		}
		inclusion := pp.buildInclusion()

		clientConfigGetter := func(context string) (*rest.Config, error) {
			kubeConfig, err := pp.buildKubeconfig(ctx)
			if err != nil {
				return nil, err
			}
			configOverrides := &clientcmd.ConfigOverrides{CurrentContext: context}
			return clientcmd.NewDefaultClientConfig(*kubeConfig, configOverrides).ClientConfig()
		}

		targetContext, err := p.NewTargetContext(ctx, clientConfigGetter, pp.d.Spec.Target, "", pp.d.Spec.DryRun, pp.d.Spec.Args, false, images, inclusion, renderOutputDir)
		if err != nil {
			return err
		}
		err = targetContext.DeploymentCollection.Prepare(targetContext.K)
		if err != nil {
			return err
		}
		return cb(targetContext)
	})
}

func (pp *preparedProject) kluctlArchive(ctx context.Context) error {
	err := pp.withKluctlProject(ctx, false, func(p *kluctl_project.LoadedKluctlProject) error {
		archivePath := filepath.Join(pp.tmpDir, "archive.tar.gz")
		metadataPath := filepath.Join(pp.tmpDir, "metadata.yaml")
		err := p.WriteArchive(archivePath, false)
		if err != nil {
			return err
		}

		metadata := p.GetMetadata()
		err = yaml.WriteYamlFile(metadataPath, &metadata)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (pp *preparedProject) handleCommandResult(ctx context.Context, cmdErr error, cmdResult *types2.CommandResult, commandName string) (*kluctlv1.CommandResult, error) {
	log := ctrl.LoggerFrom(ctx)

	log.Info(fmt.Sprintf("command finished with err=%v", cmdErr))

	revision := pp.source.GetArtifact().Revision

	if cmdErr != nil {
		pp.r.event(ctx, pp.d, revision, events.EventSeverityError, fmt.Sprintf("%s failed. %s", commandName, cmdErr.Error()), nil)
		return kluctlv1.ConvertCommandResult(cmdResult), cmdErr
	}

	msg := fmt.Sprintf("%s succeeded.", commandName)
	if len(cmdResult.NewObjects) != 0 {
		msg += fmt.Sprintf(" %d new objects.", len(cmdResult.NewObjects))
	}
	if len(cmdResult.ChangedObjects) != 0 {
		msg += fmt.Sprintf(" %d changed objects.", len(cmdResult.ChangedObjects))
	}
	if len(cmdResult.HookObjects) != 0 {
		msg += fmt.Sprintf(" %d hooks run.", len(cmdResult.HookObjects))
	}
	if len(cmdResult.DeletedObjects) != 0 {
		msg += fmt.Sprintf(" %d deleted objects.", len(cmdResult.DeletedObjects))
	}
	if len(cmdResult.OrphanObjects) != 0 {
		msg += fmt.Sprintf(" %d orphan objects.", len(cmdResult.OrphanObjects))
	}
	if len(cmdResult.Errors) != 0 {
		msg += fmt.Sprintf(" %d errors.", len(cmdResult.Errors))
	}
	if len(cmdResult.Warnings) != 0 {
		msg += fmt.Sprintf(" %d warnings.", len(cmdResult.Warnings))
	}

	severity := events.EventSeverityInfo
	var err error
	if len(cmdResult.Errors) != 0 {
		severity = events.EventSeverityError
		err = fmt.Errorf("%s failed with %d errors", commandName, len(cmdResult.Errors))
	}
	pp.r.event(ctx, pp.d, revision, severity, msg, nil)

	return kluctlv1.ConvertCommandResult(cmdResult), err
}

func (pp *preparedProject) kluctlDeploy(ctx context.Context) (*kluctlv1.CommandResult, error) {
	var retCmdResult *kluctlv1.CommandResult
	err := pp.withKluctlProjectTarget(ctx, true, func(targetContext *kluctl_project.TargetContext) error {
		cmd := commands.NewDeployCommand(targetContext.DeploymentCollection)
		cmd.ForceApply = pp.d.Spec.ForceApply
		cmd.ReplaceOnError = pp.d.Spec.ReplaceOnError
		cmd.ForceReplaceOnError = pp.d.Spec.ForceReplaceOnError
		cmd.AbortOnError = pp.d.Spec.AbortOnError
		cmd.ReadinessTimeout = time.Minute * 10
		cmd.NoWait = pp.d.Spec.NoWait

		cmdResult, err := cmd.Run(ctx, targetContext.K, nil)
		retCmdResult, err = pp.handleCommandResult(ctx, err, cmdResult, "deploy")
		return err
	})
	return retCmdResult, err
}

func (pp *preparedProject) kluctlPrune(ctx context.Context) (*kluctlv1.CommandResult, error) {
	if !pp.d.Spec.Prune {
		return nil, nil
	}

	var retCmdResult *kluctlv1.CommandResult
	err := pp.withKluctlProjectTarget(ctx, true, func(targetContext *kluctl_project.TargetContext) error {
		cmd := commands.NewPruneCommand(targetContext.DeploymentCollection)
		refs, err := cmd.Run(ctx, targetContext.K)
		if err != nil {
			return err
		}
		cmdResult, err := pp.doDeleteObjects(ctx, targetContext.K, refs)
		retCmdResult, err = pp.handleCommandResult(ctx, err, cmdResult, "deploy")
		return err
	})
	return retCmdResult, err
}

func (pp *preparedProject) kluctlDelete(ctx context.Context) (*kluctlv1.CommandResult, error) {
	if !pp.d.Spec.Prune {
		return nil, nil
	}

	var retCmdResult *kluctlv1.CommandResult
	err := pp.withKluctlProjectTarget(ctx, true, func(targetContext *kluctl_project.TargetContext) error {
		cmd := commands.NewDeleteCommand(targetContext.DeploymentCollection)
		refs, err := cmd.Run(ctx, targetContext.K)
		if err != nil {
			return err
		}
		cmdResult, err := pp.doDeleteObjects(ctx, targetContext.K, refs)
		retCmdResult, err = pp.handleCommandResult(ctx, err, cmdResult, "deploy")
		return err
	})
	return retCmdResult, err
}

func (pp *preparedProject) doDeleteObjects(ctx context.Context, k *k8s2.K8sCluster, refs []k8s.ObjectRef) (*types2.CommandResult, error) {
	log := ctrl.LoggerFrom(ctx)

	var refStrs []string
	for _, ref := range refs {
		refStrs = append(refStrs, ref.String())
	}
	if len(refStrs) != 0 {
		log.Info(fmt.Sprintf("deleting (without waiting): %s", strings.Join(refStrs, ", ")))
	}

	return utils2.DeleteObjects(k, refs, false)
}
