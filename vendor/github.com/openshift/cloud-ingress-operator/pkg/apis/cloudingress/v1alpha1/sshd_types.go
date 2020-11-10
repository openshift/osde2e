package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SSHDStateType is the overall condition of the SSHD resource
type SSHDStateType string

const (
	SSHDStateError      SSHDStateType = "Error"
	SSHDStatePending    SSHDStateType = "Pending"
	SSHDStateReady      SSHDStateType = "Ready"
	SSHDStateFinalizing SSHDStateType = "Finalizing"
)

// SSHDSpec defines the desired state of SSHD
// +k8s:openapi-gen=true
type SSHDSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// DNSName is the DNS name that should point to the SSHD service load balancers, e.g. rh-ssh
	DNSName string `json:"dnsName"`

	// AllowedCIDRBlocks is the list of CIDR blocks that should be allowed to access the SSHD service
	AllowedCIDRBlocks []string `json:"allowedCIDRBlocks"`

	// Image is the URL of the SSHD container image
	Image string `json:"image"`

	// ConfigMapSelector is a label selector to isolate config maps containing SSH authorized keys
	// to be mounted into the SSHD container
	ConfigMapSelector metav1.LabelSelector `json:"configMapSelector,omitempty"`
}

// SSHDStatus defines the observed state of SSHD
// +k8s:openapi-gen=true
type SSHDStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// State is the current state of the controller
	State SSHDStateType `json:"state,omitempty"`

	// Message is a description of the current state
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SSHD is the Schema for the sshds API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sshds,scope=Namespaced
type SSHD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHDSpec   `json:"spec,omitempty"`
	Status SSHDStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SSHDList contains a list of SSHD
type SSHDList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHD `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSHD{}, &SSHDList{})
}
