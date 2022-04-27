package controllers

import (
	"context"
	"errors"
	"fmt"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	types2 "github.com/kluctl/kluctl/v2/pkg/types"
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

func (pp *preparedProject) kluctlArchive(ctx context.Context) error {
	cmd := kluctlCaller{
		workDir:     pp.sourceDir,
		kubeconfigs: []string{pp.kubeconfig},
	}
	if pp.gitSecret != nil {
		err := cmd.addGitEnv(pp.tmpDir, pp.gitSecret)
		if err != nil {
			return err
		}
	}

	archivePath := filepath.Join(pp.tmpDir, "archive.tar.gz")
	metadataPath := filepath.Join(pp.tmpDir, "metadata.yaml")

	cmd.args = append(cmd.args, "archive")
	cmd.args = append(cmd.args, "--output-archive", archivePath)
	cmd.args = append(cmd.args, "--output-metadata", metadataPath)

	_, _, err := cmd.run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pp *preparedProject) runKluctlCommandWithResult(ctx context.Context, cmd *kluctlCaller, commandName string) (*kluctlv1.CommandResult, error) {
	resultFile := filepath.Join(pp.tmpDir, "result.yaml")
	renderOutputDir := filepath.Join(pp.tmpDir, "rendered")
	metadataFile := filepath.Join(pp.tmpDir, "metadata.yaml")
	cmd.args = append(cmd.args, "--output-format", "yaml="+resultFile)
	cmd.args = append(cmd.args, "--render-output-dir", renderOutputDir)
	cmd.args = append(cmd.args, "--output-metadata", metadataFile)

	cmd.args = append(cmd.args, "--from-archive", filepath.Join(pp.tmpDir, "archive.tar.gz"))
	cmd.args = append(cmd.args, "--from-archive-metadata", filepath.Join(pp.tmpDir, "metadata.yaml"))

	var cmdResult types2.CommandResult
	_, _, cmdErr := cmd.run(ctx)
	yamlErr := yaml.ReadYamlFile(resultFile, &cmdResult)
	if yamlErr != nil && !os.IsNotExist(errors.Unwrap(yamlErr)) {
		return nil, yamlErr
	}

	revision := pp.source.GetArtifact().Revision

	if cmdErr != nil {
		pp.r.event(ctx, pp.d, revision, events.EventSeverityError, fmt.Sprintf("%s failed. %s", commandName, cmdErr.Error()), nil)
		return kluctlv1.ConvertCommandResult(&cmdResult), cmdErr
	}
	if os.IsNotExist(yamlErr) {
		err := fmt.Errorf("%s did not write result", commandName)
		pp.r.event(ctx, pp.d, revision, events.EventSeverityInfo, err.Error(), nil)
		return nil, err
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

	return kluctlv1.ConvertCommandResult(&cmdResult), nil
}

func (pp *preparedProject) kluctlDeploy(ctx context.Context) (*kluctlv1.CommandResult, error) {
	cmd := kluctlCaller{
		workDir:     pp.sourceDir,
		kubeconfigs: []string{pp.kubeconfig},
	}

	cmd.args = append(cmd.args, "deploy")
	cmd.addTargetArgs(pp.d)
	cmd.addImageArgs(pp.d)
	cmd.addApplyArgs(pp.d)
	cmd.addInclusionArgs(pp.d)
	cmd.addMiscArgs(pp.d, true, true)
	cmd.args = append(cmd.args, "--yes")

	return pp.runKluctlCommandWithResult(ctx, &cmd, "deploy")
}

func (pp *preparedProject) kluctlPrune(ctx context.Context) (*kluctlv1.CommandResult, error) {
	if !pp.d.Spec.Prune {
		return nil, nil
	}

	cmd := kluctlCaller{
		workDir:     pp.sourceDir,
		kubeconfigs: []string{pp.kubeconfig},
	}

	cmd.args = append(cmd.args, "prune")
	cmd.addTargetArgs(pp.d)
	cmd.addImageArgs(pp.d)
	cmd.addInclusionArgs(pp.d)
	cmd.addMiscArgs(pp.d, true, false)
	cmd.args = append(cmd.args, "--yes")

	return pp.runKluctlCommandWithResult(ctx, &cmd, "prune")
}

func (pp *preparedProject) kluctlDelete(ctx context.Context) (*kluctlv1.CommandResult, error) {
	if !pp.d.Spec.Prune {
		return nil, nil
	}

	cmd := kluctlCaller{
		workDir:     pp.sourceDir,
		kubeconfigs: []string{pp.kubeconfig},
	}

	cmd.args = append(cmd.args, "delete")
	cmd.addTargetArgs(pp.d)
	cmd.addImageArgs(pp.d)
	cmd.addInclusionArgs(pp.d)
	cmd.addMiscArgs(pp.d, true, false)
	cmd.args = append(cmd.args, "--yes")

	return pp.runKluctlCommandWithResult(ctx, &cmd, "delete")
}
