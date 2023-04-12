package rosaprovider

import (
	"fmt"
	"log"
	"strings"

	createAccountRoles "github.com/openshift/rosa/cmd/create/accountroles"
	rosaAws "github.com/openshift/rosa/pkg/aws"
	"github.com/openshift/rosa/pkg/logging"
)

type AccountRoles struct {
	ControlPlaneRoleARN string
	InstallerRoleARN    string
	SupportRoleARN      string
	WorkerRoleARN       string
}

// createAccountRoles will create account roles if they do not already exist
func (m *ROSAProvider) createAccountRoles(version string) error {
	var accountRoles *AccountRoles

	prefix := fmt.Sprintf("ManagedOpenShift-%s", version)

	log.Printf("Checking if account roles exist with prefix %q", prefix)

	accountRoles, err := m.getAccountRoles(prefix, version)
	if err != nil && accountRoles != nil {
		return fmt.Errorf("fetching account roles failed: %v", err)
	}

	if accountRoles == nil {
		log.Printf("Account roles do not exist with prefix %s, creating them..", prefix)

		cmd := createAccountRoles.Cmd
		cmd.SetArgs([]string{
			"--mode", "auto",
			"--prefix", prefix,
			"--version", version,
			"--yes",
		})

		err := callAndSetAWSSession(func() error {
			return cmd.Execute()
		})
		if err != nil {
			return fmt.Errorf("error creating account roles with prefix %q, %v", prefix, err)
		}

		accountRoles, err = m.getAccountRoles(prefix, version)
		if err != nil || accountRoles == nil {
			return fmt.Errorf("fetching generated account roles with prefix failed %q, %v", prefix, err)
		}

		return nil
	}

	log.Printf("Account roles already exist with prefix %q", prefix)
	return nil
}

// getAccountRoles gets exact account roles based on prefix/version provided
func (m *ROSAProvider) getAccountRoles(prefix string, version string) (*AccountRoles, error) {
	accountRoles := &AccountRoles{}

	err := callAndSetAWSSession(func() error {
		awsClient, err := rosaAws.NewClient().
			Logger(logging.NewLogger()).
			Region(m.awsRegion).
			Build()
		if err != nil {
			return fmt.Errorf("error creating rosa aws client: %v", err)
		}

		roles, err := awsClient.ListAccountRoles(version)
		if err != nil {
			return fmt.Errorf("error listing account roles: %v", err)
		}

		count := 0
		for _, role := range roles {
			if !strings.HasPrefix(role.RoleName, prefix) {
				continue
			}

			switch role.RoleType {
			case "Control plane":
				accountRoles.ControlPlaneRoleARN = role.RoleARN
				count += 1
			case "Installer":
				accountRoles.InstallerRoleARN = role.RoleARN
				count += 1
			case "Support":
				accountRoles.SupportRoleARN = role.RoleARN
				count += 1
			case "Worker":
				accountRoles.WorkerRoleARN = role.RoleARN
				count += 1
			}
		}

		if count != 4 {
			return fmt.Errorf("error one or more prefixed %q account roles does not exist: %+v", prefix, accountRoles)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return accountRoles, nil
}
