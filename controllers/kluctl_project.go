package controllers

import (
	"context"
	"fmt"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/kluctl/kluctl/v2/pkg/deployment"
	"github.com/kluctl/kluctl/v2/pkg/deployment/commands"
	utils2 "github.com/kluctl/kluctl/v2/pkg/deployment/utils"
	"github.com/kluctl/kluctl/v2/pkg/git/auth"
	"github.com/kluctl/kluctl/v2/pkg/jinja2"
	"github.com/kluctl/kluctl/v2/pkg/kluctl_project"
	types2 "github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
)

type preparedProject struct {
	r      *KluctlDeploymentReconciler
	d      kluctlv1.KluctlDeployment
	source sourcev1.Source

	tmpDir     string
	sourceDir  string
	kubeconfig string
	gitRepo    *sourcev1.GitRepository
	gitSecret  *v1.Secret

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
		return nil, fmt.Errorf("failed to create temp dir for kluctl project: %w", err)
	}
	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		_ = os.RemoveAll(tmpDir)
	}()

	// download artifact and extract files
	err = pp.r.download(source, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed download of artifact: %w", err)
	}

	// check kluctl project path exists
	dirPath, err := securejoin.SecureJoin(tmpDir, filepath.Join("source", kluctlDeployment.Spec.Path))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(dirPath); err != nil {
		return nil, fmt.Errorf("kluctlDeployment path not found: %w", err)
	}

	kubeconfig, err := pp.writeKubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	gitSecret, err := pp.getGitSecret(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get git secret: %w", err)
	}

	pp.tmpDir = tmpDir
	pp.sourceDir = dirPath
	pp.kubeconfig = kubeconfig
	pp.gitSecret = gitSecret

	err = pp.kluctlArchive(ctx)
	if err != nil {
		return nil, err
	}

	err = yaml.ReadYamlFile(filepath.Join(pp.tmpDir, "metadata.yaml"), &pp.metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata.yaml: %w", err)
	}

	for _, t := range pp.metadata.Targets {
		if t.Target.Name == kluctlDeployment.Spec.Target {
			pp.target = t.Target
			break
		}
	}
	if pp.target == nil {
		return nil, fmt.Errorf("target %s not found in kluctl project", kluctlDeployment.Spec.Target)
	}

	archiveHash, err := calcFileHash(filepath.Join(pp.tmpDir, "archive.tar.gz"))
	if err != nil {
		return nil, err
	}
	pp.targetHash = calcTargetHash(archiveHash, pp.target)

	cleanup = false
	return pp, nil
}

func (pp *preparedProject) cleanup() {
	_ = os.RemoveAll(pp.tmpDir)
}

func (pp *preparedProject) writeKubeConfig(ctx context.Context) (string, error) {
	secretName := types.NamespacedName{
		Namespace: pp.d.GetNamespace(),
		Name:      pp.d.Spec.KubeConfig.SecretRef.Name,
	}

	var secret v1.Secret
	if err := pp.r.Get(ctx, secretName, &secret); err != nil {
		return "", fmt.Errorf("unable to read KubeConfig secret '%s' error: %w", secretName.String(), err)
	}

	var kubeConfig []byte
	for k := range secret.Data {
		if k == "value" || k == "value.yaml" {
			kubeConfig = secret.Data[k]
			break
		}
	}

	if len(kubeConfig) == 0 {
		return "", fmt.Errorf("KubeConfig secret '%s' doesn't contain a 'value' key ", secretName.String())
	}

	path := filepath.Join(pp.tmpDir, "kubeconfig.yaml")
	err := os.WriteFile(path, kubeConfig, 0o600)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (pp *preparedProject) getGitSecret(ctx context.Context) (*v1.Secret, error) {
	if gitRepo, ok := pp.source.(*sourcev1.GitRepository); ok {
		if gitRepo.Spec.SecretRef == nil {
			return nil, nil
		}
		// Attempt to retrieve secret
		name := types.NamespacedName{
			Namespace: pp.d.GetNamespace(),
			Name:      gitRepo.Spec.SecretRef.Name,
		}
		var secret v1.Secret
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

func (pp *preparedProject) buildImages() (*deployment.Images, error) {
	images, err := deployment.NewImages(pp.d.Spec.UpdateImages)
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

func (pp *preparedProject) withKluctlProject(ctx context.Context, fromArchive bool, cb func(p *kluctl_project.KluctlProjectContext) error) error {
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
	return pp.withKluctlProject(ctx, fromArchive, func(p *kluctl_project.KluctlProjectContext) error {
		renderOutputDir, err := os.MkdirTemp(pp.tmpDir, "render-")
		if err != nil {
			return err
		}
		images, err := pp.buildImages()
		if err != nil {
			return err
		}
		inclusion := pp.buildInclusion()

		targetContext, err := p.NewTargetContext(pp.d.Spec.Target, "", pp.d.Spec.DryRun, pp.d.Spec.Args, false, images, inclusion, renderOutputDir)
		if err != nil {
			return err
		}
		return cb(targetContext)
	})
}

func (pp *preparedProject) kluctlArchive(ctx context.Context) error {
	err := pp.withKluctlProject(ctx, false, func(p *kluctl_project.KluctlProjectContext) error {
		archivePath := filepath.Join(pp.tmpDir, "archive.tar.gz")
		metadataPath := filepath.Join(pp.tmpDir, "metadata.yaml")
		err := p.CreateTGZArchive(archivePath, false)
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

	pp.r.event(ctx, pp.d, revision, events.EventSeverityInfo, msg, nil)

	return kluctlv1.ConvertCommandResult(cmdResult), nil
}

func (pp *preparedProject) kluctlDeploy(ctx context.Context) (*kluctlv1.CommandResult, error) {
	var retCmdResult *kluctlv1.CommandResult
	err := pp.withKluctlProjectTarget(ctx, true, func(targetContext *kluctl_project.TargetContext) error {
		cmd := commands.NewDeployCommand(targetContext.DeploymentCollection)
		cmd.ForceApply = pp.d.Spec.ForceApply
		cmd.ReplaceOnError = pp.d.Spec.ReplaceOnError
		cmd.ForceReplaceOnError = pp.d.Spec.ForceReplaceOnError
		cmd.AbortOnError = pp.d.Spec.AbortOnError
		cmd.HookTimeout = pp.d.GetTimeout()
		cmd.NoWait = pp.d.Spec.NoWait

		cmdResult, err := cmd.Run(targetContext.K, nil)
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
		refs, err := cmd.Run(targetContext.K)
		if err != nil {
			return err
		}
		cmdResult, err := utils2.DeleteObjects(targetContext.K, refs, true)
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
		refs, err := cmd.Run(targetContext.K)
		if err != nil {
			return err
		}
		cmdResult, err := utils2.DeleteObjects(targetContext.K, refs, true)
		retCmdResult, err = pp.handleCommandResult(ctx, err, cmdResult, "deploy")
		return err
	})
	return retCmdResult, err
}
