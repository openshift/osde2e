package rosaprovider

import (
	"fmt"
	"strings"

	"github.com/openshift/osde2e/pkg/common/aws"
	createOIDCConfig "github.com/openshift/rosa/cmd/create/oidcconfig"
	deleteOIDCConfig "github.com/openshift/rosa/cmd/dlt/oidcconfig"
)

// createOIDCConfig will create a OIDC config to be used during cluster creation
func (m *ROSAProvider) createOIDCConfig(prefix string) (string, error) {
	cmd := createOIDCConfig.Cmd
	// TODO: Get installer role ARN, prompts for one when multiple exist,
	//	lets just automate the create/delete of account roles
	cmd.SetArgs([]string{
		"--mode", "auto",
		"--prefix", prefix,
		"--installer-role-arn", fmt.Sprintf("arn:aws:iam::%s:role/ManagedOpenShift-Installer-Role", aws.CcsAwsSession.GetAccount()),
		"--yes",
	})
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
func (m *ROSAProvider) deleteOIDCConfig(prefix string) error {
	oidcConfigID, err := m.getOIDCConfigID(prefix)
	if err != nil {
		return fmt.Errorf("unable to locate oidc config with prefix: %s", prefix)
	}

	cmd := deleteOIDCConfig.Cmd
	cmd.SetArgs([]string{
		"--mode", "auto",
		"--oidc-config-id", oidcConfigID,
		"--yes",
	})
	err = callAndSetAWSSession(func() error {
		return cmd.Execute()
	})
	if err != nil {
		return fmt.Errorf("error deleting odic-config ID: %s, %v", oidcConfigID, err)
	}
	return nil
}
