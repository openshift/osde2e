package spi

import (
	"fmt"
)

// ProviderCreateFunction is a function that creates providers.
type ProviderCreateFunction func() (Provider, error)

type providerRegistry struct {
	providerCreation map[string]ProviderCreateFunction
}

var registry = providerRegistry{
	providerCreation: map[string]ProviderCreateFunction{},
}

// RegisterProvider will register a provider with the given name that will be created by the given provider factory.
func RegisterProvider(name string, providerCreate func() (Provider, error)) {
	if _, ok := registry.providerCreation[name]; ok {
		panic(fmt.Sprintf("Duplicate provider name %s!", name))
	}

	registry.providerCreation[name] = providerCreate
}

// GetProvider will retrieve a provider with the given name.
func GetProvider(name string) (Provider, error) {
	if providerCreate, ok := registry.providerCreation[name]; ok {
		return providerCreate()
	}

	return nil, fmt.Errorf("unable to find provider %s", name)
}
