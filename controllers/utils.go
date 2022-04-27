package controllers

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"github.com/kluctl/kluctl/v2/pkg/types"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"github.com/kluctl/kluctl/v2/pkg/yaml"
	"io"
	"os"
)

func calcTargetHash(projectHash string, target *types.Target) string {
	s, err := yaml.WriteYamlBytes(target)
	if err != nil {
		panic(err)
	}

	h := sha256.New()
	h.Write([]byte(projectHash))
	h.Write(s)

	return hex.EncodeToString(h.Sum(nil))
}

func calcFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func calcDirHash(root string) (string, error) {
	h := sha256.New()
	tw := tar.NewWriter(h)

	err := utils.AddToTar(tw, root, "", nil)
	if err != nil {
		return "", err
	}

	err = tw.Close()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
