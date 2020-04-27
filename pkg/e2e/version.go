package e2e

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/provisioners"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	// NoVersionFound when no version can be found.
	NoVersionFound = "NoVersionFound"
)

func filterOnCincinnati(installVersion *semver.Version, upgradeVersion *semver.Version) bool {
	versionInCincinnati, err := upgrade.DoesEdgeExistInCincinnati(installVersion, upgradeVersion)

	if err != nil {
		log.Printf("error while trying to filter on version in Cincinnati: %v", err)
		return false
	}

	return versionInCincinnati
}

func removeDefaultVersion(versions []spi.Version) []spi.Version {
	versionsWithoutDefault := []spi.Version{}

	for _, version := range versions {
		if !version.Default {
			versionsWithoutDefault = append(versionsWithoutDefault, version)
		}
	}

	return versionsWithoutDefault
}

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions() (err error) {
	state := state.Instance

	// when defined, use set version
	if provisioner == nil {
		err = errors.New("osd must be setup when upgrading with release stream")
	} else if shouldUpgrade() {
		err = setupUpgradeVersion()
	} else {
		err = setupVersion()
	}

	// Set the versions in metadata. If upgrade hasn't been chosen, it should still be omitted from the end result.
	metadata.Instance.SetClusterVersion(state.Cluster.Version)
	metadata.Instance.SetUpgradeVersion(state.Upgrade.ReleaseName)

	return err
}

// shouldUpgrade determines if this test run should attempt an upgrade.
func shouldUpgrade() bool {
	cfg := config.Instance
	state := state.Instance

	return state.Upgrade.Image == "" &&
		(cfg.Upgrade.ReleaseStream != "" ||
			cfg.Upgrade.UpgradeToCISIfPossible ||
			cfg.Upgrade.NextReleaseAfterProdDefaultForUpgrade > -1)
}

// chooses between default version and nightly based on target versions.
func setupVersion() (err error) {
	cfg := config.Instance
	state := state.Instance

	versionType := "user supplied version"

	if len(state.Cluster.Version) == 0 {
		var err error

		availableVersions, err := provisioner.AvailableVersions()

		if err != nil {
			return fmt.Errorf("error getting versions: %v", err)
		}

		var selectedVersion *semver.Version
		if cfg.Cluster.UseLatestVersionForInstall {
			selectedVersion = availableVersions[len(availableVersions)-1].Version
			versionType = "latest version"
		} else if cfg.Cluster.UseMiddleClusterImageSetForInstall {
			versionsWithoutDefault := removeDefaultVersion(availableVersions)
			numVersions := len(versionsWithoutDefault)
			if numVersions < 2 {
				state.Cluster.EnoughVersionsForOldestOrMiddleTest = false
			} else {
				selectedVersion = versionsWithoutDefault[numVersions/2].Version
			}
			versionType = "middle version"
		} else if cfg.Cluster.UseOldestClusterImageSetForInstall {
			versionsWithoutDefault := removeDefaultVersion(availableVersions)
			numVersions := len(versionsWithoutDefault)
			if numVersions < 2 {
				state.Cluster.EnoughVersionsForOldestOrMiddleTest = false
			} else {
				selectedVersion = versionsWithoutDefault[0].Version
			}
			versionType = "oldest version"
		} else if cfg.Cluster.NextReleaseAfterProdDefault > -1 {
			var prodDefault *semver.Version
			prodDefault, err = getProdDefault()

			if err == nil {
				selectedVersion, err = nextReleaseAfterGivenVersionFromVersionList(prodDefault, availableVersions, cfg.Cluster.NextReleaseAfterProdDefault)
				versionType = fmt.Sprintf("%d release(s) from the default version in prod", cfg.Cluster.NextReleaseAfterProdDefault)
			}
		} else if cfg.OCM.Env == "int" {
			selectedVersion, err = getProdDefault()
			versionType = "current default in prod"
		} else {
			for _, version := range availableVersions {
				if version.Default {
					selectedVersion = version.Version
				}
			}
			if selectedVersion == nil {
				err = fmt.Errorf("unable to find default version")
			}
			versionType = "current default"
		}

		if err == nil {
			if state.Cluster.EnoughVersionsForOldestOrMiddleTest {
				state.Cluster.Version = util.SemverToOpenshiftVersion(selectedVersion)
			} else {
				log.Printf("Unable to get the %s.", versionType)
			}
		} else {
			return fmt.Errorf("error finding default cluster version: %v", err)
		}
	} else {
		// Make sure the cluster version is valid
		_, err := util.OpenshiftVersionToSemver(state.Cluster.Version)

		if err != nil {
			return fmt.Errorf("supplied version %s is invalid: %v", state.Cluster.Version, err)
		}
	}

	log.Printf("Using the %s '%s'", versionType, state.Cluster.Version)

	return
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion() (err error) {
	cfg := config.Instance
	state := state.Instance

	// Decide the version to install
	err = setupVersion()
	if err != nil {
		return err
	}

	clusterVersion, err := util.OpenshiftVersionToSemver(state.Cluster.Version)
	if err != nil {
		log.Printf("error while parsing cluster version %s: %v", state.Cluster.Version, err)
		return err
	}

	if cfg.OCM.Env != "int" {
		availableVersions, err := provisioner.AvailableVersions()

		if err != nil {
			return err
		}

		filteredVersionList := []*semver.Version{}

		for _, version := range availableVersions {
			if filterOnCincinnati(clusterVersion, version.Version) {
				filteredVersionList = append(filteredVersionList, version.Version)
			}
		}

		numFilteredVersions := len(filteredVersionList)

		if numFilteredVersions == 0 {
			log.Printf("no edges found for install version %s", state.Cluster.Version)
			state.Upgrade.ReleaseName = NoVersionFound
			return nil
		}

		cisUpgradeVersionString := util.SemverToOpenshiftVersion(filteredVersionList[len(filteredVersionList)-1])

		if cisUpgradeVersionString == NoVersionFound {
			state.Upgrade.ReleaseName = cisUpgradeVersionString
			metadata.Instance.SetUpgradeVersionSource("none")
			return nil
		}

		cisUpgradeVersion, err := util.OpenshiftVersionToSemver(cisUpgradeVersionString)

		if err != nil {
			log.Printf("unable to parse most recent version of openshift from OSD: %v", err)
			return err
		}

		// If the available cluster image set makes sense, then we'll just use that
		if !cisUpgradeVersion.LessThan(clusterVersion) {
			log.Printf("Using cluster image set.")
			state.Upgrade.ReleaseName = cisUpgradeVersionString
			metadata.Instance.SetUpgradeVersionSource("cluster image set")
			state.Upgrade.UpgradeVersionEqualToInstallVersion = cisUpgradeVersion.Equal(clusterVersion)
			log.Printf("Selecting version '%s' to be able to upgrade to '%s'", state.Cluster.Version, state.Upgrade.ReleaseName)
			return nil
		}

		if state.Upgrade.ReleaseName != "" {
			log.Printf("The most recent cluster image set is equal to the default. Falling back to upgrading with Cincinnati.")
		} else {
			return fmt.Errorf("couldn't get latest cluster image set release and no Cincinnati fallback")
		}
	}

	releaseStream := cfg.Upgrade.ReleaseStream

	if releaseStream == "" {
		if cfg.Upgrade.NextReleaseAfterProdDefaultForUpgrade > -1 {
			availableVersions, err := provisioner.AvailableVersions()

			if err != nil {
				return fmt.Errorf("error getting available versions: %v", err)
			}

			prodDefault, err := getProdDefault()

			if err != nil {
				return fmt.Errorf("error getting production default: %v", err)
			}

			nextVersion, err := nextReleaseAfterGivenVersionFromVersionList(prodDefault, availableVersions, cfg.Upgrade.NextReleaseAfterProdDefaultForUpgrade)

			if err != nil {
				return fmt.Errorf("error determining next version to upgrade to: %v", err)
			}

			releaseStream = fmt.Sprintf("%d.%d.0-0.nightly", nextVersion.Major(), nextVersion.Minor())
		} else {
			return fmt.Errorf("no release stream specified and no dynamic version selection specified")
		}
	}

	state.Upgrade.ReleaseName, state.Upgrade.Image, err = upgrade.LatestReleaseFromReleaseController(releaseStream)
	if err != nil {
		return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
	}

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' on release stream '%s'",
		state.Cluster.Version, state.Upgrade.ReleaseName, releaseStream)
	return
}

func getProdDefault() (*semver.Version, error) {
	prodProvisioner, err := provisioners.ClusterProvisionerForProduction()

	if err != nil {
		return nil, fmt.Errorf("unable to get production provisioner for getting versions from production: %v", err)
	}

	versions, err := prodProvisioner.AvailableVersions()

	if err != nil {
		return nil, fmt.Errorf("unable to get versions from production: %v", err)
	}

	for _, version := range versions {
		if version.Default {
			return version.Version, nil
		}
	}

	return nil, fmt.Errorf("no default found on prod")
}

// nextReleaseAfterGivenVersionFromVersionList will attempt to look for the next valid X.Y stream release, given a delta (releaseFromGivenVersion)
// Example In/Out
// In: 4.3.12, [4.3.13, 4.4.0, 4.5.0], 2
// Out: 4.5.0, nil
func nextReleaseAfterGivenVersionFromVersionList(givenVersion *semver.Version, versionList []spi.Version, releasesFromGivenVersion int) (*semver.Version, error) {
	versionBuckets := map[string]*semver.Version{}

	// Assemble a map that lists a release (x.y.0) to its latest version, with nightlies taking precedence over all else
	for _, version := range versionList {
		versionSemver := version.Version
		majorMinor := createMajorMinorStringFromSemver(versionSemver)

		if _, ok := versionBuckets[majorMinor]; !ok {
			versionBuckets[majorMinor] = versionSemver
		} else {
			currentGreatestVersion := versionBuckets[majorMinor]
			versionIsNightly := strings.Contains(versionSemver.Prerelease(), "nightly")
			currentIsNightly := strings.Contains(currentGreatestVersion.Prerelease(), "nightly")

			// Make sure nightlies take precedence over other versions
			if versionIsNightly && !currentIsNightly {
				versionBuckets[majorMinor] = versionSemver
			} else if currentIsNightly && !versionIsNightly {
				continue
			} else if currentGreatestVersion.LessThan(versionSemver) {
				versionBuckets[majorMinor] = versionSemver
			}
		}
	}

	// Parse all major minor versions (x.y.0) into semver versions and place them in an array.
	// This is done explicitly so that we can utilize the semver library's sorting capability.
	majorMinorList := []*semver.Version{}
	for k := range versionBuckets {
		parsedMajorMinor, err := semver.NewVersion(k)
		if err != nil {
			return nil, err
		}

		majorMinorList = append(majorMinorList, parsedMajorMinor)
	}

	sort.Sort(semver.Collection(majorMinorList))

	// Now that the list is sorted, we want to locate the major minor of the given version in the list.
	givenMajorMinor, err := semver.NewVersion(createMajorMinorStringFromSemver(givenVersion))

	if err != nil {
		return nil, err
	}

	indexOfGivenMajorMinor := -1
	for i, majorMinor := range majorMinorList {
		if majorMinor.Equal(givenMajorMinor) {
			indexOfGivenMajorMinor = i
			break
		}
	}

	if indexOfGivenMajorMinor == -1 {
		return nil, fmt.Errorf("unable to find current prod default in %s environment", config.Instance.OCM.Env)
	}

	// Next, we'll go the given version distance ahead of the given version. We want to do it this way instead of guessing
	// the next minor release so that we can handle major releases in the future, In other words, if the Openshift
	// 4.y line stops at 4.13, we'll still be able to pick 5.0 if it's the next release after 4.13.
	nextMajorMinorIndex := indexOfGivenMajorMinor + releasesFromGivenVersion

	if len(majorMinorList) <= nextMajorMinorIndex {
		return nil, fmt.Errorf("there is no eligible next release on the %s environment", config.Instance.OCM.Env)
	}
	nextMajorMinor := createMajorMinorStringFromSemver(majorMinorList[nextMajorMinorIndex])

	if _, ok := versionBuckets[nextMajorMinor]; !ok {
		return nil, fmt.Errorf("no major/minor version found for %s", nextMajorMinor)
	}

	return versionBuckets[nextMajorMinor], nil
}

func createMajorMinorStringFromSemver(version *semver.Version) string {
	return fmt.Sprintf("%d.%d", version.Major(), version.Minor())
}
