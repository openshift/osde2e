package installselectors

// registry is the install version selector registry.
type registry struct {
	selectors []Interface
}

var globalSelectorRegistry registry = registry{
	selectors: []Interface{},
}

// GetVersionSelectors will return the registered version selectors for initial cluster installation.
func GetVersionSelectors() []Interface {
	return globalSelectorRegistry.selectors
}

func registerSelector(i Interface) {
	globalSelectorRegistry.selectors = append(globalSelectorRegistry.selectors, i)
}
