package ocmprovisioner

import (
	"fmt"
	"log"
	"sort"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-"

	// PageSize is the number of results to get per page from the cluster versions endpoint
	PageSize = 100

	// NoVersionFound is the value placed into a version string when no valid Cincinnati version can be selected.
	NoVersionFound = "NoVersionFound"
)

// AvailableVersions will return all of the available version.
func (o *OCMProvisioner) AvailableVersions() ([]spi.Version, error) {
	var err error

	o.versionCacheOnce.Do(func() {
		versions := []spi.Version{}

		page := 1
		log.Printf("Querying cluster versions endpoint.")
		for {
			var resp *v1.VersionsListResponse
			err = retryer().Do(func() error {
				var err error

				resp, err = o.conn.ClustersMgmt().V1().Versions().List().Page(page).Size(PageSize).Send()

				if err != nil {
					return err
				}

				if resp != nil && resp.Error() != nil {
					return errResp(resp.Error())
				}

				return nil
			})

			if err != nil {
				err = fmt.Errorf("failed getting list of OSD versions: %v", err)
			} else if resp != nil {
				err = errResp(resp.Error())
			}

			if err != nil {
				log.Print("error getting cluster versions from getSemverList.Response")
				log.Printf("Response Headers: %v", resp.Header())
				log.Printf("Response Error(s): %v", resp.Error())
				log.Printf("HTTP Code: %d", resp.Status())
				log.Printf("Size of response: %d", resp.Size())

				err = fmt.Errorf("couldn't retrieve available versions: %v", err)
				return
			}

			// parse versions, filter for major+minor nightlies, then sort
			resp.Items().Each(func(v *v1.Version) bool {
				if version, err := util.OpenshiftVersionToSemver(v.ID()); err != nil {
					log.Printf("could not parse version '%s': %v", v.ID(), err)
				} else if v.Enabled() {
					versions = append(versions, spi.Version{
						Version: version,
						Default: v.Default(),
					})
				}
				return true
			})

			// If we've looked at all the results, stop collecting them.
			if page*PageSize >= resp.Total() {
				break
			}
			page++
		}

		sort.Slice(versions, func(i, j int) bool {
			return versions[i].Version.LessThan(versions[j].Version)
		})

		o.versionCache = versions
	})

	if err != nil {
		return nil, err
	}

	if o.versionCache == nil {
		return nil, fmt.Errorf("error getting versions, please refer to log for details")
	}

	return o.versionCache, nil
}
