package ocmprovider

import (
	"errors"
	"fmt"
	"log"

	accounts "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

// CheckQuota determines if enough quota is available to launch with cfg.
func (o *OCMProvider) CheckQuota(skuRuleID string) (bool, error) {
	// get flavour being deployed
	var skuResp *accounts.SkuRuleGetResponse
	err := retryer().Do(func() error {
		var err error
		if skuRuleID == "" {
			return fmt.Errorf("No valid SKU selected")
		}
		skuResp, err = o.conn.AccountsMgmt().V1().SkuRules().SkuRule(skuRuleID).Get().Send()

		if err != nil {
			return err
		}

		if skuResp != nil && skuResp.Error() != nil {
			err = errResp(skuResp.Error())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("error trying to get SKUs: %v", err)
	}

	if skuResp == nil || skuResp.Body().Empty() {
		return false, errors.New("returned SKU can't be empty")
	}
	sku := skuResp.Body()

	// get quota
	quotaList, err := o.currentAccountQuota()
	if err != nil {
		return false, fmt.Errorf("could not get quota: %v", err)
	}

	skuQuota, hasQuota := sku.GetQuotaId()
	if !hasQuota {
		// Assume this is a bad thing? All SKU rules currently have a quota associated.
		return false, errors.New("returned SKU has no associated quota")
	}

	quotaFound := false
	for _, q := range quotaList.Slice() {
		if quotaFound = HasQuotaForSKU(q, skuQuota); quotaFound {
			log.Printf("Quota for test config (sku=%s/quota=%s/multiAZ=%t) found: total=%d, remaining: %d",
				skuRuleID, skuQuota, viper.GetBool(config.Cluster.MultiAZ), q.Allowed(), q.Allowed()-q.Consumed())
			break
		}
	}

	return quotaFound, nil
}

// CurrentAccountQuota returns quota available for the current account's organization in the environment.
func (o *OCMProvider) currentAccountQuota() (*accounts.QuotaCostList, error) {
	resp, err := o.conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil || resp == nil {
		return nil, fmt.Errorf("couldn't get current account: %v", err)
	}

	acc := resp.Body()

	if acc.Organization() == nil || acc.Organization().ID() == "" {
		return nil, fmt.Errorf("organization for account '%s' must be set to get quota", acc.ID())
	}

	orgID := acc.Organization().ID()

	var quotaList *accounts.QuotaCostListResponse
	err = retryer().Do(func() error {
		var err error
		quotaList, err = o.conn.AccountsMgmt().V1().Organizations().Organization(orgID).QuotaCost().List().Send()

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

// HasQuotaForSKU looks for a quota cost for the desired SKU and returns it if one is found
// and sufficient quota exists.
func HasQuotaForSKU(q *accounts.QuotaCost, skuQuota string) bool {
	if q.QuotaID() == skuQuota {
		if q.Consumed() < q.Allowed() {
			return true
		}
	}
	return false
}
