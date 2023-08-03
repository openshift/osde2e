package rosaprovider

import (
	"fmt"
	"strings"

	createOIDCConfig "github.com/openshift/rosa/cmd/create/oidcconfig"
	deleteOIDCConfig "github.com/openshift/rosa/cmd/dlt/oidcconfig"
)

// createOIDCConfig will create a OIDC config to be used during cluster creation
func (m *ROSAProvider) createOIDCConfig(prefix, installerRoleARN string) (string, error) {
	cmd := createOIDCConfig.Cmd
	args := []string{
		"--mode", "auto",
		"--prefix", prefix,
		"--managed=false",
		"--installer-role-arn", installerRoleARN,
		"--yes",
	}

	cmd.SetArgs(args)

	err := callAndSetAWSSession(func() error {
		return cmd.Execute()
	})
	if err != nil {
		return "", fmt.Errorf("error creating odic-config: %v", err)
	}

	oidcConfigID, err := m.getOIDCConfigID(prefix)
	if err != nil {
		return "", fmt.Errorf("OIDC config created, %v", err)
	}

	return oidcConfigID, nil
}

// getOIDCConfigID gets an existing oidc config id
func (m *ROSAProvider) getOIDCConfigID(prefix string) (string, error) {
	response, err := m.ocmProvider.GetConnection().ClustersMgmt().V1().OidcConfigs().List().Send()
	if err != nil {
		return "", fmt.Errorf("unable to get oidc configs from ocm: %v", err)
	}

	for _, oidcConfig := range response.Items().Slice() {
		if strings.Contains(oidcConfig.SecretArn(), prefix) {
			return oidcConfig.ID(), nil
		}
	}

	return "", fmt.Errorf("unable to locate oidc config with prefix: %s", prefix)
}

// deleteOIDCConfig deletes an existing oidc config that was used for cluster creation
func (m *ROSAProvider) deleteOIDCConfig(oidcConfigID string) error {
	cmd := deleteOIDCConfig.Cmd
	cmd.SetArgs([]string{
		"--mode", "auto",
		"--interactive=false",
		"--oidc-config-id", oidcConfigID,
		"--yes",
	})
	err := callAndSetAWSSession(func() error {
		return cmd.Execute()
	})
	if err != nil {
		return fmt.Errorf("error deleting odic-config ID: %s, %v", oidcConfigID, err)
	}
	return nil
}
