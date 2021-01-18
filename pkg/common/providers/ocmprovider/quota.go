package ocmprovider

import (
	"errors"
	"fmt"
	"log"

	accounts "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

const (
	// resourceClusterFmt is the format string for a quota resource type for a cluster.
	resourceClusterFmt = "cluster.%s"
)

// CheckQuota determines if enough quota is available to launch with cfg.
func (o *OCMProvider) CheckQuota(flavourID string) (bool, error) {
	// get flavour being deployed
	var flavourResp *v1.FlavourGetResponse
	err := retryer().Do(func() error {
		var err error
		if flavourID == "" {
			return fmt.Errorf("No valid flavour selected")
		}
		flavourResp, err = o.conn.ClustersMgmt().V1().Flavours().Flavour(flavourID).Get().Send()

		if err != nil {
			return err
		}

		if flavourResp != nil && flavourResp.Error() != nil {
			err = errResp(flavourResp.Error())
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, fmt.Errorf("error trying to get flavours: %v", err)
	}

	if flavourResp == nil || flavourResp.Body().Empty() {
		return false, errors.New("returned flavour can't be empty")
	}
	flavour := flavourResp.Body()

	// get quota
	quotaList, err := o.currentAccountQuota()
	if err != nil {
		return false, fmt.Errorf("could not get quota: %v", err)
	}

	// TODO: use compute_machine_type when available in OCM SDK
	_ = flavour.Nodes()
	machineType := ""

	quotaFound := false
	resourceClusterType := fmt.Sprintf(resourceClusterFmt, viper.GetString(config.CloudProvider.CloudProviderID))
	for _, q := range quotaList.Slice() {
		if quotaFound = HasQuotaFor(q, resourceClusterType, machineType); quotaFound {
			log.Printf("Quota for test config (%s/%s/multiAZ=%t) found: total=%d, remaining: %d",
				resourceClusterType, machineType, viper.GetBool(config.Cluster.MultiAZ), q.Allowed(), q.Allowed()-q.Reserved())
			break
		}
	}

	return quotaFound, nil
}

// CurrentAccountQuota returns quota available for the current account's organization in the environment.
func (o *OCMProvider) currentAccountQuota() (*accounts.QuotaSummaryList, error) {
	resp, err := o.conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil || resp == nil {
		return nil, fmt.Errorf("couldn't get current account: %v", err)
	}

	acc := resp.Body()

	if acc.Organization() == nil || acc.Organization().ID() == "" {
		return nil, fmt.Errorf("organization for account '%s' must be set to get quota", acc.ID())
	}

	orgID := acc.Organization().ID()

	var quotaList *accounts.QuotaSummaryListResponse
	err = retryer().Do(func() error {
		var err error
		quotaList, err = o.conn.AccountsMgmt().V1().Organizations().Organization(orgID).QuotaSummary().List().Send()

		if err != nil {
			return err
		}

		if quotaList != nil && quotaList.Error() != nil {
			return errResp(quotaList.Error())
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting quota list: %v", err)
	} else if quotaList == nil {
		return nil, errors.New("QuotaList can't be nil")
	}

	return quotaList.Items(), nil
}

// HasQuotaFor the desired configuration. If machineT is empty a default will try to be selected.
func HasQuotaFor(q *accounts.QuotaSummary, resourceType, machineType string) bool {
	azType := "single"
	if viper.GetBool(config.Cluster.MultiAZ) {
		azType = "multi"
	}

	if q.ResourceType() == resourceType && q.ResourceName() == machineType || machineType == "" {
		if q.AvailabilityZoneType() == azType || q.AvailabilityZoneType() == "any" {
			if q.Reserved() < q.Allowed() {
				return true
			}
		}
	}
	return false
}
