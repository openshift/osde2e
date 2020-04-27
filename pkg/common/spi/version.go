package spi

import "github.com/Masterminds/semver"

// Version represents an Openshift version.
type Version struct {
	version   *semver.Version
	isDefault bool
}

// Version is the actual version found by the provider.
func (v *Version) Version() *semver.Version {
	return v.version
}

// Default is whether or not the version is the default for the provider.
func (v *Version) Default() bool {
	return v.isDefault
}

// VersionBuilder is used to build version objects.
type VersionBuilder struct {
	version   *semver.Version
	isDefault bool
}

// NewVersionBuilder creates a new version builder.
func NewVersionBuilder() *VersionBuilder {
	return &VersionBuilder{}
}

// Version sets the version for the builder.
func (vb *VersionBuilder) Version(version *semver.Version) *VersionBuilder {
	vb.version = version
	return vb
}

// Default sets the isDefault value for the builder.
func (vb *VersionBuilder) Default(isDefault bool) *VersionBuilder {
	vb.isDefault = isDefault
	return vb
}

// Build will build the version object.
func (vb *VersionBuilder) Build() *Version {
	return &Version{
		version:   vb.version,
		isDefault: vb.isDefault,
	}
}

// VersionList is the list of versions found by the provider.
type VersionList struct {
	availableVersions      []*Version
	defaultVersionOverride *semver.Version
}

// AvailableVersions is the list of versions available to the provider.
func (vl *VersionList) AvailableVersions() []*Version {
	return vl.availableVersions
}

// Default is the default version in the VersionList. If defaultVersionOverride is
// set, it will be returned instead of the availableVersion tagged as default.
func (vl *VersionList) Default() *semver.Version {
	if vl.defaultVersionOverride != nil {
		return vl.defaultVersionOverride
	}

	for _, version := range vl.availableVersions {
		if version.isDefault {
			return version.version
		}
	}

	return nil
}

// VersionListBuilder will build VersionList objects.
type VersionListBuilder struct {
	availableVersions      []*Version
	defaultVersionOverride *semver.Version
}

// NewVersionListBuilder creates a new version list builder.
func NewVersionListBuilder() *VersionListBuilder {
	return &VersionListBuilder{}
}

// AvailableVersions sets the available versions in the builder.
func (vb *VersionListBuilder) AvailableVersions(availableVersions []*Version) *VersionListBuilder {
	vb.availableVersions = availableVersions
	return vb
}

// DefaultVersionOverride sets the default version override in the builder.
func (vb *VersionListBuilder) DefaultVersionOverride(defaultVersionOverride *semver.Version) *VersionListBuilder {
	vb.defaultVersionOverride = defaultVersionOverride
	return vb
}

// Build will build the version list object.
func (vb *VersionListBuilder) Build() *VersionList {
	return &VersionList{
		availableVersions:      vb.availableVersions,
		defaultVersionOverride: vb.defaultVersionOverride,
	}
}
