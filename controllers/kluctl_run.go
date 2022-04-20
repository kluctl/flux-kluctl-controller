package controllers

import (
	"bufio"
	"bytes"
	"fmt"
	kluctlv1 "github.com/kluctl/flux-kluctl-controller/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type kluctlCaller struct {
	workDir            string
	localClusters      *string
	localDeployment    *string
	localSealedSecrets *string
	kubeconfigs        []string

	args []string
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

func (kc *kluctlCaller) run() (string, string, error) {
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

	env := os.Environ()
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", strings.Join(kc.kubeconfigs, string(os.PathListSeparator))))

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
