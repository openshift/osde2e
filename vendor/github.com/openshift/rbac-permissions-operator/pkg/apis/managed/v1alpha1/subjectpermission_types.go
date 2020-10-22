package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SubjectPermissionSpec defines the desired state of SubjectPermission
// +k8s:openapi-gen=true
type SubjectPermissionSpec struct {
	// Kind of the Subject that is being granted permissions by the operator
	SubjectKind string `json:"subjectKind"`
	// Name of the Subject granted permissions by the operator
	SubjectName string `json:"subjectName"`
	// List of permissions applied at Cluster scope
	// +optional
	ClusterPermissions []string `json:"clusterPermissions,omitempty"`
	// List of permissions applied at Namespace scope
	// +optional
	Permissions []Permission `json:"permissions,omitempty"`
}

// Permission defines a Role that is bound to the Subject
// Allowed in specific Namespaces
type Permission struct {
	// ClusterRoleName to bind to the Subject as a RoleBindings in allowed Namespaces
	ClusterRoleName string `json:"clusterRoleName"`
	// NamespacesAllowedRegex representing allowed Namespaces
	NamespacesAllowedRegex string `json:"namespacesAllowedRegex,omitempty"`
	// NamespacesDeniedRegex representing denied Namespaces
	NamespacesDeniedRegex string `json:"namespacesDeniedRegex,omitempty"`
	// Flag to indicate if "allow" regex is applied first
	// If 'true' order is Allow then Deny, Else order is Deny then Allow
	AllowFirst bool `json:"allowFirst"`
}

// SubjectPermissionStatus defines the observed state of SubjectPermission
// +k8s:openapi-gen=true
type SubjectPermissionStatus struct {
	// List of conditions for the CR
	Conditions []Condition `json:"conditions,omitempty"`
	// State that this condition represents
	State string `json:"state"`
}

// Condition defines a single condition of running the operator against an instance of the SubjectPermission CR
type Condition struct {
	// Type is the type of the condition
	Type SubjectPermissionType `json:"type,omitempty"`
	// LastTransitionTime is the last time this condition was active for the CR
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Message related to the condition
	// +optional
	Message string `json:"message,omitempty"`
	// ClusterRoleName in which this condition is true
	ClusterRoleNames []string `json:"clusterRoleName,omitempty"`
	// Flag to indicate if condition status is currently active
	Status bool `json:"status"`
	// State that this condition represents
	State SubjectPermissionState `json:"state"`
}

// SubjectPermissionState defines various states a SubjectPermission CR can be in
type SubjectPermissionState string

// SubjectPermissionType defines various type a SubjectPermission CR can be in
type SubjectPermissionType string

const (
	// ClusterRoleBindingCreated const for ClusterRoleBindingCreated status
	ClusterRoleBindingCreated SubjectPermissionType = "ClusterRoleBindingCreated"
	// RoleBindingCreated const for RoleBindingCreated status
	RoleBindingCreated SubjectPermissionType = "RoleBindingCreated"
	// SubjectPermissionStateCreated const for Created state
	SubjectPermissionStateCreated SubjectPermissionState = "Created"
	// SubjectPermissionStateFailed const for Failed state
	SubjectPermissionStateFailed SubjectPermissionState = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubjectPermission is the Schema for the subjectpermissions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SubjectPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubjectPermissionSpec   `json:"spec,omitempty"`
	Status SubjectPermissionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubjectPermissionList contains a list of SubjectPermission
type SubjectPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubjectPermission `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubjectPermission{}, &SubjectPermissionList{})
}
