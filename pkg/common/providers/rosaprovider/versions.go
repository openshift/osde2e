package rosaprovider

import (
	"context"
	"fmt"
	"log"

	"github.com/Masterminds/semver/v3"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
)

// Versions returns a list of enabled rosa versions for cluster creation
func (m *ROSAProvider) Versions() (*spi.VersionList, error) {
	var (
		err                    error
		ctx                                    = context.Background()
		spiVersions                            = []*spi.Version{}
		defaultVersionOverride *semver.Version = nil
	)

	availableVersions, err := m.provider.Versions(
		ctx,
		viper.GetString(config.Cluster.Channel),
		viper.GetBool(config.Hypershift),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve rosa versions: %v", err)
	}

	for _, version := range availableVersions {
		if semVersion, err := util.OpenshiftVersionToSemver(version.ID); err != nil {
			log.Printf("could not parse version '%s': %v", version.ID, err)
		} else if version.Enabled {
			if version.Default {
				defaultVersionOverride = semVersion
			}

			spiVersion := spi.NewVersionBuilder().
				Version(semVersion).
				Default(version.Default).
				Build()

			for _, upgrade := range version.AvailableUpgrades {
				if version, err := util.OpenshiftVersionToSemver(upgrade); err == nil {
					spiVersion.AddUpgradePath(version)
				}
			}

			spiVersions = append(spiVersions, spiVersion)
		}
	}

	return spi.NewVersionListBuilder().
			AvailableVersions(spiVersions).
			DefaultVersionOverride(defaultVersionOverride).
			Build(),
		nil
}
