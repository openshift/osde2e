package customerrors

// ProvisionClusterError defines a collection of fields for the error type
type ProvisionClusterError struct {
	Created bool
	Message string
}

// Custom error for cluster provision errors
func (e *ProvisionClusterError) Error() string {
	return e.Message
}

// ClusterHealthCheckError defines a collection of fields for the error type
type ClusterHealthCheckError struct {
	Message string
}

// Error implementation for custom cluster health check error type
func (e *ClusterHealthCheckError) Error() string {
	return e.Message
}

// KubeConfigLookupError defines a collection of fields for the error type
type KubeConfigLookupError struct {
	Message string
}

// Error implementation for custom kubeconfig look up error type
func (e *KubeConfigLookupError) Error() string {
	return e.Message
}

// AddOnInstallError defines a collection of fields for the error type
type AddOnInstallError struct {
	Message string
}

// Error implementation for custom addon installation error type
func (e *AddOnInstallError) Error() string {
	return e.Message
}
