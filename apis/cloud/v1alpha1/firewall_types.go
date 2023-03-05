/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type FirewallResource struct {
	// +kubebuilder:validation:Enum=server;label_selector
	Type string `json:"type"`

	// +optional
	LabelSelector *string `json:"label_selector,omitempty"`

	// +optional
	Server *int `json:"server,omitempty"`
}

type FirewallRule struct {
	// +optional
	Description *string `json:"description,omitempty"`

	// +optional
	SourceIPs []string `json:"source_ips,omitempty"`

	// +optional
	DestinationIPs []string `json:"destination_ips,omitempty"`

	// +kubebuilder:validation:Enum=in;out
	Direction string `json:"direction"`

	// +kubebuilder:validation:Enum=tcp;udp;icmp;esp;gre
	Protocol string `json:"protocol"`

	// +optional
	Port *string `json:"port"`
}

// FirewallParameters are the configurable fields of a Firewall.
type FirewallParameters struct {
	// +optional
	Rules []FirewallRule `json:"rules,omitempty"`

	// +optional
	ApplyTo []FirewallResource `json:"apply_to,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// FirewallObservation are the observable fields of a Firewall.
type FirewallObservation struct {
	Id      int          `json:"id"`
	Created *metav1.Time `json:"created,omitempty"`
}

// A FirewallSpec defines the desired state of a Firewall.
type FirewallSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       FirewallParameters `json:"forProvider"`
}

// A FirewallStatus represents the observed state of a Firewall.
type FirewallStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          FirewallObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Firewall is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,hetzner}
type Firewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallSpec   `json:"spec"`
	Status FirewallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallList contains a list of Firewall
type FirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Firewall `json:"items"`
}

// Firewall type metadata.
var (
	FirewallKind             = reflect.TypeOf(Firewall{}).Name()
	FirewallGroupKind        = schema.GroupKind{Group: Group, Kind: FirewallKind}.String()
	FirewallKindAPIVersion   = FirewallKind + "." + SchemeGroupVersion.String()
	FirewallGroupVersionKind = SchemeGroupVersion.WithKind(FirewallKind)
)

func init() {
	SchemeBuilder.Register(&Firewall{}, &FirewallList{})
}
