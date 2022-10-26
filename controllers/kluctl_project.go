package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_jinja2"
	"github.com/kluctl/kluctl/v2/pkg/utils/uo"
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
	"github.com/kluctl/kluctl/v2/pkg/git/repocache"
	k8s2 "github.com/kluctl/kluctl/v2/pkg/k8s"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	"github.com/kluctl/kluctl/v2/pkg/registries"
	types2 "github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/types/k8s"
	"github.com/kluctl/kluctl/v2/pkg/utils"
)

type preparedProject struct {
	r      *KluctlDeploymentReconciler
	obj    *kluctlv1.KluctlDeployment
	source sourcev1.Source

	tmpDir     string
	repoDir    string
	projectDir string

	gitRepo *sourcev1.GitRepository
}

type preparedTarget struct {
	pp *preparedProject
}

func prepareProject(ctx context.Context,
	r *KluctlDeploymentReconciler,
	obj *kluctlv1.KluctlDeployment,
	source sourcev1.Source) (*preparedProject, error) {

	pp := &preparedProject{
		r:      r,
		obj:    obj,
		source: source,
	}

	// create tmp dir
	tmpDir, err := os.MkdirTemp("", obj.GetName()+"-")
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

	if source != nil {
		// download artifact and extract files
		err = pp.r.download(source, tmpDir)
		if err != nil {
			return pp, fmt.Errorf("failed download of artifact: %w", err)
		}

		pp.repoDir = filepath.Join(tmpDir, "source")

		// check kluctl project path exists
		pp.projectDir, err = securejoin.SecureJoin(pp.repoDir, pp.obj.Spec.Path)
		if err != nil {
			return pp, err
		}
		if _, err := os.Stat(pp.projectDir); err != nil {
			return pp, fmt.Errorf("kluctlDeployment path not found: %w", err)
		}
	}

	cleanup = false
	return pp, nil
}

func (pp *preparedProject) cleanup() {
	_ = os.RemoveAll(pp.tmpDir)
}

func (pp *preparedProject) newTarget() (*preparedTarget, error) {
	pt := preparedTarget{
		pp: pp,
	}

	return &pt, nil
}

func (pt *preparedTarget) restConfigToKubeconfig(restConfig *rest.Config) *api.Config {
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
	user.Exec = restConfig.ExecProvider
	kubeConfig.AuthInfos["default"] = user

	kctx := api.NewContext()
	kctx.Cluster = "default"
	kctx.AuthInfo = "default"
	kubeConfig.Contexts["default"] = kctx

	return kubeConfig
}

func (pt *preparedTarget) getKubeconfigFromSecret(ctx context.Context) ([]byte, error) {
	secretName := types.NamespacedName{
		Namespace: pt.pp.obj.GetNamespace(),
		Name:      pt.pp.obj.Spec.KubeConfig.SecretRef.Name,
	}

	var secret corev1.Secret
	if err := pt.pp.r.Get(ctx, secretName, &secret); err != nil {
		return nil, fmt.Errorf("unable to read KubeConfig secret '%s' error: %w", secretName.String(), err)
	}

	var kubeConfig []byte
	switch {
	case pt.pp.obj.Spec.KubeConfig.SecretRef.Key != "":
		key := pt.pp.obj.Spec.KubeConfig.SecretRef.Key
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

func (pt *preparedTarget) setImpersonationConfig(restConfig *rest.Config) {
	name := pt.pp.r.DefaultServiceAccount
	if sa := pt.pp.obj.Spec.ServiceAccountName; sa != "" {
		name = sa
	}
	if name != "" {
		username := fmt.Sprintf("system:serviceaccount:%s:%s", pt.pp.obj.GetNamespace(), name)
		restConfig.Impersonate = rest.ImpersonationConfig{UserName: username}
	}
}

func (pt *preparedTarget) buildRestConfig(ctx context.Context) (*rest.Config, error) {
	var err error
	var restConfig *rest.Config

	if pt.pp.obj.Spec.KubeConfig != nil {
		kubeConfig, err := pt.getKubeconfigFromSecret(ctx)
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

	pt.setImpersonationConfig(restConfig)

	return restConfig, nil
}

func (pt *preparedTarget) buildKubeconfig(ctx context.Context) (*api.Config, error) {
	restConfig, err := pt.buildRestConfig(ctx)
	if err != nil {
		return nil, err
	}

	kubeConfig := pt.restConfigToKubeconfig(restConfig)

	err = pt.renameContexts(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

func (pt *preparedTarget) renameContexts(cfg *api.Config) error {
	for _, r := range pt.pp.obj.Spec.RenameContexts {
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

func (pt *preparedTarget) getRegistrySecrets(ctx context.Context) ([]*corev1.Secret, error) {
	var ret []*corev1.Secret
	for _, ref := range pt.pp.obj.Spec.RegistrySecrets {
		name := types.NamespacedName{
			Namespace: pt.pp.obj.GetNamespace(),
			Name:      ref.Name,
		}
		var secret corev1.Secret
		if err := pt.pp.r.Get(ctx, name, &secret); err != nil {
			return nil, fmt.Errorf("failed to get secret '%s': %w", name.String(), err)
		}
		ret = append(ret, &secret)
	}
	return ret, nil
}

func (pt *preparedTarget) buildRegistryHelper(ctx context.Context) (*registries.RegistryHelper, error) {
	secrets, err := pt.getRegistrySecrets(ctx)
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

func (pt *preparedTarget) buildImages(ctx context.Context) (*deployment.Images, error) {
	rh, err := pt.buildRegistryHelper(ctx)
	if err != nil {
		return nil, err
	}
	offline := !pt.pp.obj.Spec.UpdateImages
	images, err := deployment.NewImages(rh, pt.pp.obj.Spec.UpdateImages, offline)
	if err != nil {
		return nil, err
	}
	for _, fi := range kluctlv1.ConvertFixedImagesToKluctl(pt.pp.obj.Spec.Images) {
		images.AddFixedImage(fi)
	}
	return images, nil
}

func (pt *preparedTarget) buildInclusion() *utils.Inclusion {
	inc := utils.NewInclusion()
	for _, x := range pt.pp.obj.Spec.IncludeTags {
		inc.AddInclude("tag", x)
	}
	for _, x := range pt.pp.obj.Spec.ExcludeTags {
		inc.AddExclude("tag", x)
	}
	for _, x := range pt.pp.obj.Spec.IncludeDeploymentDirs {
		inc.AddInclude("deploymentItemDir", x)
	}
	for _, x := range pt.pp.obj.Spec.ExcludeDeploymentDirs {
		inc.AddExclude("deploymentItemDir", x)
	}
	return inc
}

func (pt *preparedTarget) clientConfigGetter(ctx context.Context) func(context *string) (*rest.Config, *api.Config, error) {
	return func(context *string) (*rest.Config, *api.Config, error) {
		kubeConfig, err := pt.buildKubeconfig(ctx)
		if err != nil {
			return nil, nil, err
		}

		configOverrides := &clientcmd.ConfigOverrides{}
		if context != nil {
			configOverrides.CurrentContext = *context
		}
		clientConfig := clientcmd.NewDefaultClientConfig(*kubeConfig, configOverrides)
		rawConfig, err := clientConfig.RawConfig()
		if err != nil {
			return nil, nil, err
		}
		if context != nil {
			rawConfig.CurrentContext = *context
		}
		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return nil, nil, err
		}
		return restConfig, &rawConfig, nil
	}
}

func (pp *preparedProject) withKluctlProject(ctx context.Context, pt *preparedTarget, cb func(p *kluctl_project.LoadedKluctlProject) error) error {
	j2, err := kluctl_jinja2.NewKluctlJinja2(true)
	if err != nil {
		return err
	}
	defer j2.Close()

	ga, err := pp.r.buildGitAuth(ctx, pp.source, pp.obj.GetNamespace())
	if err != nil {
		return err
	}

	rp := repocache.NewGitRepoCache(ctx, pp.r.SshPool, ga, nil, 0)
	defer rp.Clear()

	loadArgs := kluctl_project.LoadKluctlProjectArgs{
		RepoRoot:   pp.repoDir,
		ProjectDir: pp.projectDir,
		RP:         rp,
	}
	if pt != nil {
		loadArgs.ClientConfigGetter = pt.clientConfigGetter(ctx)
	}

	p, err := kluctl_project.LoadKluctlProject(ctx, loadArgs, filepath.Join(pp.tmpDir, "project"), j2)
	if err != nil {
		return err
	}

	return cb(p)
}

func (pp *preparedProject) listTargets(ctx context.Context) ([]*types2.Target, error) {
	var ret []*types2.Target
	err := pp.withKluctlProject(ctx, nil, func(p *kluctl_project.LoadedKluctlProject) error {
		for _, x := range p.DynamicTargets {
			ret = append(ret, x.Target)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (pt *preparedTarget) withKluctlProjectTarget(ctx context.Context, cb func(targetContext *kluctl_project.TargetContext) error) error {
	return pt.pp.withKluctlProject(ctx, pt, func(p *kluctl_project.LoadedKluctlProject) error {
		renderOutputDir, err := os.MkdirTemp(pt.pp.tmpDir, "render-")
		if err != nil {
			return err
		}
		images, err := pt.buildImages(ctx)
		if err != nil {
			return err
		}
		inclusion := pt.buildInclusion()

		externalArgs, err := uo.FromString(string(pt.pp.obj.Spec.Args.Raw))
		if err != nil {
			return err
		}
		props := kluctl_project.TargetContextParams{
			TargetName:      pt.pp.obj.Spec.Target,
			DryRun:          pt.pp.obj.Spec.DryRun,
			ExternalArgs:    externalArgs,
			Images:          images,
			Inclusion:       inclusion,
			RenderOutputDir: renderOutputDir,
		}
		targetContext, err := p.NewTargetContext(ctx, props)
		if err != nil {
			return err
		}
		err = targetContext.DeploymentCollection.Prepare()
		if err != nil {
			return err
		}
		return cb(targetContext)
	})
}

func (pt *preparedTarget) handleCommandResult(ctx context.Context, cmdErr error, cmdResult *types2.CommandResult, commandName string) error {
	log := ctrl.LoggerFrom(ctx)

	log.Info(fmt.Sprintf("command finished with err=%v", cmdErr))

	kluctlv1.RemoveObjectsFromCommandResult(cmdResult)

	revision := ""
	if pt.pp.source != nil {
		revision = pt.pp.source.GetArtifact().Revision
	}

	if cmdErr != nil {
		pt.pp.r.event(ctx, pt.pp.obj, revision, events.EventSeverityError, fmt.Sprintf("%s failed. %s", commandName, cmdErr.Error()), nil)
		return cmdErr
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
	pt.pp.r.event(ctx, pt.pp.obj, revision, severity, msg, nil)

	return err
}

func (pt *preparedTarget) kluctlDeploy(ctx context.Context, targetContext *kluctl_project.TargetContext) (*types2.CommandResult, error) {
	cmd := commands.NewDeployCommand(targetContext.DeploymentCollection)
	cmd.ForceApply = pt.pp.obj.Spec.ForceApply
	cmd.ReplaceOnError = pt.pp.obj.Spec.ReplaceOnError
	cmd.ForceReplaceOnError = pt.pp.obj.Spec.ForceReplaceOnError
	cmd.AbortOnError = pt.pp.obj.Spec.AbortOnError
	cmd.ReadinessTimeout = time.Minute * 10
	cmd.NoWait = pt.pp.obj.Spec.NoWait

	cmdResult, err := cmd.Run(ctx, targetContext.SharedContext.K, nil)
	err = pt.handleCommandResult(ctx, err, cmdResult, "deploy")
	return cmdResult, err
}

func (pt *preparedTarget) kluctlPokeImages(ctx context.Context, targetContext *kluctl_project.TargetContext) (*types2.CommandResult, error) {
	cmd := commands.NewPokeImagesCommand(targetContext.DeploymentCollection)

	cmdResult, err := cmd.Run(ctx, targetContext.SharedContext.K)
	err = pt.handleCommandResult(ctx, err, cmdResult, "poke-images")
	return cmdResult, err
}

func (pt *preparedTarget) kluctlPrune(ctx context.Context, targetContext *kluctl_project.TargetContext) (*types2.CommandResult, error) {
	cmd := commands.NewPruneCommand(targetContext.DeploymentCollection)
	refs, err := cmd.Run(ctx, targetContext.SharedContext.K)
	if err != nil {
		return nil, err
	}
	cmdResult, err := pt.doDeleteObjects(ctx, targetContext.SharedContext.K, refs)
	err = pt.handleCommandResult(ctx, err, cmdResult, "prune")
	return cmdResult, err
}

func (pt *preparedTarget) kluctlValidate(ctx context.Context, targetContext *kluctl_project.TargetContext) (*types2.ValidateResult, error) {
	cmd := commands.NewValidateCommand(ctx, targetContext.DeploymentCollection)

	cmdResult, err := cmd.Run(ctx, targetContext.SharedContext.K)
	return cmdResult, err
}

func (pt *preparedTarget) kluctlDelete(ctx context.Context, commonLabels map[string]string) (*types2.CommandResult, error) {
	if !pt.pp.obj.Spec.Prune {
		return nil, nil
	}

	cmd := commands.NewDeleteCommand(nil)
	cmd.OverrideDeleteByLabels = commonLabels

	restConfig, err := pt.buildRestConfig(ctx)
	if err != nil {
		return nil, err
	}
	clientFactory, err := k8s2.NewClientFactory(restConfig)
	if err != nil {
		return nil, err
	}
	k, err := k8s2.NewK8sCluster(ctx, clientFactory, pt.pp.obj.Spec.DryRun)
	if err != nil {
		return nil, err
	}

	refs, err := cmd.Run(ctx, k)
	if err != nil {
		return nil, err
	}
	cmdResult, err := pt.doDeleteObjects(ctx, k, refs)
	err = pt.handleCommandResult(ctx, err, cmdResult, "delete")
	return cmdResult, err
}

func (pt *preparedTarget) doDeleteObjects(ctx context.Context, k *k8s2.K8sCluster, refs []k8s.ObjectRef) (*types2.CommandResult, error) {
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
