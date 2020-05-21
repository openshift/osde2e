package upgradeselectors

// registry is the upgrade version selector registry.
type registry struct {
	selectors []Interface
}

var globalSelectorRegistry registry = registry{
	selectors: []Interface{},
}

// GetVersionSelectors will return the registered version selectors for cluster upgrades.
func GetVersionSelectors() []Interface {
	return globalSelectorRegistry.selectors
}

func registerSelector(u Interface) {
	globalSelectorRegistry.selectors = append(globalSelectorRegistry.selectors, u)
}
