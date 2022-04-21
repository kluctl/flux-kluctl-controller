package controllers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"io"
	corev1 "k8s.io/api/core/v1"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

type kluctlCaller struct {
	workDir            string
	localClusters      *string
	localDeployment    *string
	localSealedSecrets *string
	kubeconfigs        []string

	args []string
	env  []string

	tmpFiles []string
}

func (kc *kluctlCaller) deleteTmpFiles() {
	for _, f := range kc.tmpFiles {
		_ = os.Remove(f)
	}
	kc.tmpFiles = nil
}

func (kc *kluctlCaller) addTargetArgs(kluctlDeployment kluctlv1.KluctlDeployment) {
	kc.args = append(kc.args, "--target", kluctlDeployment.Spec.Target)

	for k, v := range kluctlDeployment.Spec.Args {
		kc.args = append(kc.args, "-a", fmt.Sprintf("%s=%s", k, v))
	}
}

func (kc *kluctlCaller) addImageArgs(kluctlDeployment kluctlv1.KluctlDeployment) {
	if kluctlDeployment.Spec.UpdateImages {
		kc.args = append(kc.args, "-u")
	}
}

func (kc *kluctlCaller) addMiscArgs(kluctlDeployment kluctlv1.KluctlDeployment, dryRun bool, wait bool) {
	if dryRun && kluctlDeployment.Spec.DryRun {
		kc.args = append(kc.args, "--dry-run")
	}
	if wait && kluctlDeployment.Spec.NoWait {
		kc.args = append(kc.args, "--no-wait")
	}
	if wait {
		kc.args = append(kc.args, "--hook-timeout", kluctlDeployment.GetTimeout().String())
	}
}

func (kc *kluctlCaller) addApplyArgs(kluctlDeployment kluctlv1.KluctlDeployment) {
	if kluctlDeployment.Spec.ForceApply {
		kc.args = append(kc.args, "--force-apply")
	}
	if kluctlDeployment.Spec.ReplaceOnError {
		kc.args = append(kc.args, "--replace-on-error")
	}
	if kluctlDeployment.Spec.ForceReplaceOnError {
		kc.args = append(kc.args, "--force-replace-on-error")
	}
}

func (kc *kluctlCaller) addInclusionArgs(kluctlDeployment kluctlv1.KluctlDeployment) {
	for _, x := range kluctlDeployment.Spec.IncludeTags {
		kc.args = append(kc.args, "--include-tag", x)
	}
	for _, x := range kluctlDeployment.Spec.ExcludeTags {
		kc.args = append(kc.args, "--exclude-tag", x)
	}
	for _, x := range kluctlDeployment.Spec.IncludeDeploymentDirs {
		kc.args = append(kc.args, "--include-deployment-dir", x)
	}
	for _, x := range kluctlDeployment.Spec.ExcludeDeploymentDirs {
		kc.args = append(kc.args, "--exclude-deployment-dir", x)
	}
}

func (kc *kluctlCaller) addGitEnv(tmpDir string, u url.URL, secret *corev1.Secret) error {
	writeEnvFile := func(secretKey string, envName string) error {
		x, ok := secret.Data[secretKey]
		if !ok {
			return nil
		}

		tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("ssh-%s-", secretKey))
		if err != nil {
			return err
		}
		_, err = tmpFile.Write(x)
		if err != nil {
			return fmt.Errorf("failed to write temporary file: %w", err)
		}
		_ = tmpFile.Close()
		kc.tmpFiles = append(kc.tmpFiles, tmpFile.Name())
		kc.env = append(kc.env, fmt.Sprintf("%s=%s", envName, tmpFile.Name()))
		return nil
	}

	kc.env = append(kc.env, fmt.Sprintf("KLUCTL_GIT_HOST=%s", u.Hostname()))
	if x, ok := secret.Data["username"]; ok {
		kc.env = append(kc.env, fmt.Sprintf("KLUCTL_GIT_USERNAME=%s", string(x)))
	} else if u.User != nil && u.User.Username() != "" {
		kc.env = append(kc.env, fmt.Sprintf("KLUCTL_GIT_USERNAME=%s", u.User.Username()))
	}
	if x, ok := secret.Data["password"]; ok {
		kc.env = append(kc.env, fmt.Sprintf("KLUCTL_GIT_PASSWORD=%s", string(x)))
	}
	err := writeEnvFile("caFile", "KLUCTL_GIT_CA_FILE")
	if err != nil {
		return err
	}
	err = writeEnvFile("identity", "KLUCTL_GIT_SSH_KEY")
	if err != nil {
		return err
	}
	err = writeEnvFile("known_hosts", "SSH_KNOWN_HOSTS")
	if err != nil {
		return err
	}
	return nil
}

func (kc *kluctlCaller) run(ctx context.Context) (string, string, error) {
	log := ctrl.LoggerFrom(ctx)

	var args []string
	args = append(args, kc.args...)
	args = append(args, "--no-update-check")

	if kc.localClusters != nil {
		args = append(args, "--local-clusters", *kc.localClusters)
	}
	if kc.localDeployment != nil {
		args = append(args, "--local-deployment", *kc.localDeployment)
	}
	if kc.localSealedSecrets != nil {
		args = append(args, "--local-sealed-secrets", *kc.localSealedSecrets)
	}

	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "KUBECONFIG=") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", strings.Join(kc.kubeconfigs, string(os.PathListSeparator))))
	env = append(env, kc.env...)

	kluctlExe := os.Getenv("KLUCTL_EXE")
	if kluctlExe == "" {
		curDir, _ := os.Getwd()
		for i, p := range env {
			x := strings.SplitN(p, "=", 2)
			if x[0] == "PATH" {
				env[i] = fmt.Sprintf("PATH=%s%c%s%c%s", curDir, os.PathListSeparator, filepath.Join(curDir, ".."), os.PathListSeparator, x[1])
			}
		}
		kluctlExe = "kluctl"
	} else {
		p, err := filepath.Abs(kluctlExe)
		if err != nil {
			return "", "", err
		}
		kluctlExe = p
	}

	log.Info(fmt.Sprintf("calling kluctl with args: %s", strings.Join(args, " ")))

	cmd := exec.Command(kluctlExe, args...)
	cmd.Dir = kc.workDir
	cmd.Env = env

	stdout, stderr, err := runHelper(cmd)
	return stdout, stderr, err
}

func runHelper(cmd *exec.Cmd) (string, string, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		_ = stdoutPipe.Close()
		return "", "", err
	}

	stdReader := func(testLogPrefix string, buf io.StringWriter, pipe io.Reader) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			l := scanner.Text()
			logrus.Infof(testLogPrefix + l)
			_, _ = buf.WriteString(l + "\n")
		}
	}

	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)

	go stdReader("stdout: ", stdoutBuf, stdoutPipe)
	go stdReader("stderr: ", stderrBuf, stderrPipe)

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
