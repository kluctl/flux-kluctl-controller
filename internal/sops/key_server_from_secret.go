package sops

import (
	"bytes"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	intkeyservice "github.com/kluctl/flux-kluctl-controller/internal/sops/keyservice"
	"go.mozilla.org/sops/v3/age"
	"go.mozilla.org/sops/v3/azkv"
	"go.mozilla.org/sops/v3/hcvault"
	"go.mozilla.org/sops/v3/keyservice"
	"go.mozilla.org/sops/v3/kms"
	"go.mozilla.org/sops/v3/pgp"
	corev1 "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
)

const (
	// DecryptionPGPExt is the extension of the file containing an armored PGP
	// key.
	DecryptionPGPExt = ".asc"
	// DecryptionAgeExt is the extension of the file containing an age key
	// file.
	DecryptionAgeExt = ".agekey"
	// DecryptionVaultTokenFileName is the name of the file containing the
	// Hashicorp Vault token.
	DecryptionVaultTokenFileName = "sops.vault-token"
	// DecryptionAWSKmsFile is the name of the file containing the AWS KMS
	// credentials.
	DecryptionAWSKmsFile = "sops.aws-kms"
	// DecryptionAzureAuthFile is the name of the file containing the Azure
	// credentials.
	DecryptionAzureAuthFile = "sops.azure-kv"
	// DecryptionGCPCredsFile is the name of the file containing the GCP
	// credentials.
	DecryptionGCPCredsFile = "sops.gcp-kms"
)

func BuildSopsKeyServerFromSecret(secret *corev1.Secret, gnuPGHomeDir string, opts ...intkeyservice.ServerOption) (keyservice.KeyServiceClient, error) {
	gnuPGHome := pgp.GnuPGHome(gnuPGHomeDir)

	var ageIdentities age.ParsedIdentities
	var vaultToken hcvault.Token
	var awsCredsProvider *kms.CredentialsProvider
	var azureToken azcore.TokenCredential
	var gcpCredsJSON []byte

	var err error

	for name, value := range secret.Data {
		switch filepath.Ext(name) {
		case DecryptionPGPExt:
			if err = gnuPGHome.Import(value); err != nil {
				return nil, fmt.Errorf("failed to import '%s' data from decryption Secret: %w", name, err)
			}
		case DecryptionAgeExt:
			if err = ageIdentities.Import(string(value)); err != nil {
				return nil, fmt.Errorf("failed to import '%s' data from decryption Secret: %w", name, err)
			}
		case filepath.Ext(DecryptionVaultTokenFileName):
			// Make sure we have the absolute name
			if name == DecryptionVaultTokenFileName {
				token := string(value)
				token = strings.Trim(strings.TrimSpace(token), "\n")
				vaultToken = hcvault.Token(token)
			}
		case filepath.Ext(DecryptionAWSKmsFile):
			if name == DecryptionAWSKmsFile {
				if awsCredsProvider, err = LoadCredsProviderFromYaml(value); err != nil {
					return nil, fmt.Errorf("failed to import data from decryption Secret '%s': %w", name, err)
				}
			}
		case filepath.Ext(DecryptionAzureAuthFile):
			// Make sure we have the absolute name
			if name == DecryptionAzureAuthFile {
				conf := AADConfig{}
				if err = LoadAADConfigFromBytes(value, &conf); err != nil {
					return nil, fmt.Errorf("failed to import '%s' data from decryption Secret: %w", name, err)
				}
				if azureToken, err = TokenFromAADConfig(conf); err != nil {
					return nil, fmt.Errorf("failed to import '%s' data from decryption Secret: %w", name, err)
				}
			}
		case filepath.Ext(DecryptionGCPCredsFile):
			if name == DecryptionGCPCredsFile {
				gcpCredsJSON = bytes.Trim(value, "\n")
			}
		}
	}

	serverOpts := []intkeyservice.ServerOption{
		intkeyservice.WithGnuPGHome(gnuPGHome),
		intkeyservice.WithVaultToken(vaultToken),
		intkeyservice.WithAgeIdentities(ageIdentities),
		intkeyservice.WithGCPCredsJSON(gcpCredsJSON),
	}
	serverOpts = append(serverOpts, opts...)
	if azureToken != nil {
		serverOpts = append(serverOpts, intkeyservice.WithAzureToken{Token: azkv.NewTokenCredential(azureToken)})
	}
	serverOpts = append(serverOpts, intkeyservice.WithAWSKeys{CredsProvider: awsCredsProvider})
	server := intkeyservice.NewServer(serverOpts...)

	return keyservice.NewCustomLocalClient(server), nil
}
