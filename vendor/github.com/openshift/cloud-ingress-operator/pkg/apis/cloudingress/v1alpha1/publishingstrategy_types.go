package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PublishingStrategySpec defines the desired state of PublishingStrategy
type PublishingStrategySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// DefaultAPIServerIngress defines whether API is internal or external
	DefaultAPIServerIngress DefaultAPIServerIngress `json:"defaultAPIServerIngress"`
	//ApplicationIngress defines whether application ingress is internal or external
	ApplicationIngress []ApplicationIngress `json:"applicationIngress"`
}

// DefaultAPIServerIngress defines API ingress
type DefaultAPIServerIngress struct {
	// Listening defines internal or external ingress
	Listening Listening `json:"listening,omitempty"`
}

// ApplicationIngress defines application ingress
type ApplicationIngress struct {
	// Listening defines application ingress as internal or external
	Listening Listening `json:"listening,omitempty"`
	// Default defines default value of ingress when cluster installs
	Default       bool                   `json:"default"`
	DNSName       string                 `json:"dnsName"`
	Certificate   corev1.SecretReference `json:"certificate"`
	RouteSelector metav1.LabelSelector   `json:"routeSelector,omitempty"`
}

// Listening defines internal or external api and ingress
type Listening string

const (
	// Internal const for listening status
	Internal Listening = "internal"
	// External const for listening status
	External Listening = "external"
)

// PublishingStrategyStatus defines the observed state of PublishingStrategy
type PublishingStrategyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PublishingStrategy is the Schema for the publishingstrategies API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=publishingstrategies,scope=Namespaced
type PublishingStrategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublishingStrategySpec   `json:"spec"`
	Status PublishingStrategyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PublishingStrategyList contains a list of PublishingStrategy
type PublishingStrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublishingStrategy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PublishingStrategy{}, &PublishingStrategyList{})
}
