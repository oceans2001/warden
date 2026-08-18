// +kubebuilder:object:generate=true
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AgentSandboxSpec struct {
	Image          string   `json:"image"`
	Command        []string `json:"command,omitempty"`
	AllowedDomains []string `json:"allowedDomains"`
}

type AgentSandboxStatus struct {
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type AgentSandbox struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSandboxSpec   `json:"spec,omitempty"`
	Status AgentSandboxStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AgentSandboxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AgentSandbox `json:"items"`
}
