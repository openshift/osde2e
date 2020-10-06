package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APISchemeConditionType - APISchemeConditionType
type APISchemeConditionType string

const (
	ConditionError APISchemeConditionType = "Error"
	ConditionReady APISchemeConditionType = "Ready"
)

// APISchemeSpec defines the desired state of APIScheme
// +k8s:openapi-gen=true
type APISchemeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ManagementAPIServerIngress ManagementAPIServerIngress `json:"managementAPIServerIngress"`
}

// ManagementAPIServerIngress defines the Management API ingress
type ManagementAPIServerIngress struct {
	// Enabled to create the Management API endpoint or not.
	Enabled bool `json:"enabled"`
	// DNSName is the name that should be used for DNS of the management API, eg rh-api
	DNSName string `json:"dnsName"`
	// AllowedCIDRBlocks is the list of CIDR blocks that should be allowed to access the management API
	AllowedCIDRBlocks []string `json:"allowedCIDRBlocks"`
}

// APISchemeStatus defines the observed state of APIScheme
// +k8s:openapi-gen=true
type APISchemeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	CloudLoadBalancerDNSName string                 `json:"cloudLoadBalancerDNSName,omitempty"`
	Conditions               []APISchemeCondition   `json:"conditions,omitempty"`
	State                    APISchemeConditionType `json:"state,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIScheme is the Schema for the APISchemes API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=APISchemes,scope=Namespaced
type APIScheme struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APISchemeSpec   `json:"spec"`
	Status APISchemeStatus `json:"status,omitempty"`
}

// APISchemeCondition is the history of transitions
type APISchemeCondition struct {
	// Type is the type of condition
	Type APISchemeConditionType `json:"type,omitempty"`

	// LastTransitionTime Last change to status
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// LastProbeTime last time probed
	LastProbeTime metav1.Time `json:"lastProbeTime"`

	// AllowedCIDRBlocks currently allowed (as of the last successful Security Group update)
	AllowedCIDRBlocks []string `json:"allowedCIDRBlocks,omitempty"`

	// Reason is why we're making this status change
	Reason string `json:"reason"`

	// Message is an English text
	Message string `json:"message"`
	// Status
	Status corev1.ConditionStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APISchemeList contains a list of APIScheme
type APISchemeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIScheme `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIScheme{}, &APISchemeList{})
}
