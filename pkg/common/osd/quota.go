package osd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"

	accounts "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	osderrors "github.com/openshift-online/ocm-sdk-go/errors"

	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	// ResourceAWSCluster is the quota resource type for a cluster on AWS.
	ResourceAWSCluster = "cluster.aws"
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
	for _, q := range quotaList {
		if quotaFound = HasQuotaFor(q, ResourceAWSCluster, machineType); quotaFound {
			log.Printf("Quota for test config (%s/%s/multiAZ=%t) found: total=%d, remaining: %d",
				ResourceAWSCluster, machineType, config.Instance.Cluster.MultiAZ, q.Allowed(), q.Allowed()-q.Reserved())
		}
	}

	return quotaFound, nil
}

// CurrentAccountQuota returns quota available for the current account's organization in the environment.
func (u *OSD) CurrentAccountQuota() ([]*accounts.ResourceQuota, error) {
	acc, err := u.CurrentAccount()
	if err != nil || acc == nil {
		return nil, fmt.Errorf("couldn't get current account: %v", err)
	} else if acc.Organization() == nil || acc.Organization().ID() == "" {
		return nil, fmt.Errorf("organization for account '%s' must be set to get quota", acc.ID())
	}

	orgID := acc.Organization().ID()
	quotaList, err := u.getQuotaSummary(orgID)
	if err == nil && quotaList != nil {
		err = errResp(quotaList.Error())
	} else if quotaList == nil {
		return nil, errors.New("QuotaList can't be nil")
	}
	return quotaList.Items(), err
}

// HasQuotaFor the desired configuration. If machineT is empty a default will try to be selected.
func HasQuotaFor(q *accounts.ResourceQuota, resourceType, machineType string) bool {
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

// TODO: use ocm-sdk-go resource_summary method once available
func (u *OSD) getQuotaSummary(orgID string) (*resourceSummaryListResponse, error) {
	resp := new(resourceSummaryListResponse)
	summaryPath := path.Join("/api/accounts_mgmt", APIVersion, "organizations", orgID, "quota_summary")
	rawResp, err := u.conn.Get().Path(summaryPath).Send()
	if err == nil && rawResp.Status() != http.StatusOK {
		resp.err, err = osderrors.UnmarshalError(rawResp.Bytes())
	} else if rawResp != nil {
		err = json.Unmarshal(rawResp.Bytes(), resp)
	}

	if err != nil {
		return resp, err
	}

	// allow reading QuotaSummary as ResourceQuota
	for i := range resp.List {
		resp.List[i]["kind"] = "ResourceQuota"
	}

	// convert formats by writing to bytes and unmarshalling typed
	var listData []byte
	if listData, err = json.Marshal(resp.List); err == nil {
		resp.list, err = accounts.UnmarshalResourceQuotaList(listData)
	}
	return resp, err
}

type resourceSummaryListResponse struct {
	Kind  string                   `json:"kind"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
	Total int                      `json:"total"`
	List  []map[string]interface{} `json:"items"`

	list []*accounts.ResourceQuota
	err  *osderrors.Error
}

func (r *resourceSummaryListResponse) Items() []*accounts.ResourceQuota {
	return r.list
}
func (r *resourceSummaryListResponse) Error() *osderrors.Error {
	return r.err
}
