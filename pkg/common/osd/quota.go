package osd

import (
	"errors"
	"fmt"
	"log"

	accounts "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"

	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	// ResourceClusterFmt is the format string for a quota resource type for a cluster.
	ResourceClusterFmt = "cluster.%s"
)

// CheckQuota determines if enough quota is available to launch with cfg.
func (u *OSD) CheckQuota() (bool, error) {
	// get flavour being deployed
	flavourID := u.Flavour()
	flavourReq, err := u.conn.ClustersMgmt().V1().Flavours().Flavour(flavourID).Get().Send()
	if err == nil && flavourReq != nil {
		err = errResp(flavourReq.Error())
		if err != nil {
			return false, err
		}
	} else if flavourReq == nil || flavourReq.Body().Empty() {
		return false, errors.New("returned flavour can't be empty")
	}
	flavour := flavourReq.Body()

	// get quota
	quotaList, err := u.CurrentAccountQuota()
	if err != nil {
		return false, fmt.Errorf("could not get quota: %v", err)
	}

	// TODO: use compute_machine_type when available in OCM SDK
	_ = flavour.Nodes()
	machineType := ""

	quotaFound := false
	resourceClusterType := fmt.Sprintf(ResourceClusterFmt, config.Instance.CloudProvider.CloudProviderID)
	for _, q := range quotaList.Slice() {
		if quotaFound = HasQuotaFor(q, resourceClusterType, machineType); quotaFound {
			log.Printf("Quota for test config (%s/%s/multiAZ=%t) found: total=%d, remaining: %d",
				resourceClusterType, machineType, config.Instance.Cluster.MultiAZ, q.Allowed(), q.Allowed()-q.Reserved())
			break
		}
	}

	return quotaFound, nil
}

// CurrentAccountQuota returns quota available for the current account's organization in the environment.
func (u *OSD) CurrentAccountQuota() (*accounts.QuotaSummaryList, error) {
	acc, err := u.CurrentAccount()
	if err != nil || acc == nil {
		return nil, fmt.Errorf("couldn't get current account: %v", err)
	} else if acc.Organization() == nil || acc.Organization().ID() == "" {
		return nil, fmt.Errorf("organization for account '%s' must be set to get quota", acc.ID())
	}

	orgID := acc.Organization().ID()
	quotaList, err := u.conn.AccountsMgmt().V1().Organizations().Organization(orgID).QuotaSummary().List().Send()
	if err == nil && quotaList != nil {
		err = errResp(quotaList.Error())
	} else if quotaList == nil {
		return nil, errors.New("QuotaList can't be nil")
	}

	return quotaList.Items(), err
}

// HasQuotaFor the desired configuration. If machineT is empty a default will try to be selected.
func HasQuotaFor(q *accounts.QuotaSummary, resourceType, machineType string) bool {
	azType := "single"
	if config.Instance.Cluster.MultiAZ {
		azType = "multi"
	}

	if q.ResourceType() == resourceType && q.ResourceName() == machineType || machineType == "" {
		if q.AvailabilityZoneType() == azType {
			if q.Reserved() < q.Allowed() {
				return true
			}
		}
	}
	return false
}
