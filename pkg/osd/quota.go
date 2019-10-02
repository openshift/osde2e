package osd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"

	accounts "github.com/openshift-online/uhc-sdk-go/pkg/client/accountsmgmt/v1"
	osderrors "github.com/openshift-online/uhc-sdk-go/pkg/client/errors"

	"github.com/openshift/osde2e/pkg/config"
)

const (
	// ResourceAWSCluster is the quota resource type for a cluster on AWS.
	ResourceAWSCluster = "cluster.aws"
)

// CheckQuota determines if enough quota is available to launch with cfg.
func (u *OSD) CheckQuota(cfg *config.Config) (bool, error) {
	// get flavour being deployed
	flavourId := u.Flavour(cfg)
	flavourReq, err := u.conn.ClustersMgmt().V1().Flavours().Flavour(flavourId).Get().Send()
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

	// TODO: use compute_machine_type when available in UHC SDK
	_ = flavour.Nodes()
	machineType := ""

	quotaFound := false
	quotaList.Each(func(q *accounts.ResourceQuota) bool {
		if quotaFound = HasQuotaFor(q, cfg, ResourceAWSCluster, machineType); quotaFound {
			log.Printf("Quota for test config (%s/%s/multiAZ=%t) found: total=%d, remaining: %d",
				ResourceAWSCluster, machineType, cfg.MultiAZ, q.Allowed(), q.Allowed()-q.Reserved())
		}
		return !quotaFound
	})

	return quotaFound, nil
}

// CurrentAccountQuota returns quota available for the current account's organization in the environment.
func (u *OSD) CurrentAccountQuota() (*accounts.ResourceQuotaList, error) {
	acc, err := u.CurrentAccount()
	if err != nil || acc == nil {
		return nil, fmt.Errorf("couldn't get current account: %v", err)
	} else if acc.Organization() == nil || acc.Organization().ID() == "" {
		return nil, fmt.Errorf("organization for account '%s' must be set to get quota", acc.ID())
	}

	orgId := acc.Organization().ID()
	quotaList, err := u.getQuotaSummary(orgId)
	if err == nil && quotaList != nil {
		err = errResp(quotaList.Error())
	} else if quotaList == nil {
		return nil, errors.New("QuotaList can't be nil")
	}
	return quotaList.Items(), err
}

// HasQuotaFor the desired configuration. If machineT is empty a default will try to be selected.
func HasQuotaFor(q *accounts.ResourceQuota, cfg *config.Config, resourceType, machineType string) bool {
	azType := "single"
	if cfg.MultiAZ {
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

// TODO: use uhc-sdk-go resource_summary method once available
func (u *OSD) getQuotaSummary(orgId string) (*resourceSummaryListResponse, error) {
	resp := new(resourceSummaryListResponse)
	summaryPath := path.Join("/api/accounts_mgmt", APIVersion, "organizations", orgId, "quota_summary")
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

	list *accounts.ResourceQuotaList
	err  *osderrors.Error
}

func (r *resourceSummaryListResponse) Items() *accounts.ResourceQuotaList {
	return r.list
}
func (r *resourceSummaryListResponse) Error() *osderrors.Error {
	return r.err
}
