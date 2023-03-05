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
	"k8s.io/apimachinery/pkg/util/intstr"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// PublicNetwork describes the public network to configure for a Server
type PublicNetwork struct {
	// +optional
	EnableIPv4 *bool `json:"enable_ipv4,omitempty"`
	// +optional
	EnableIPv6 *bool `json:"enable_ipv6,omitempty"`
	// +optional
	IPv4 *int `json:"ipv4,omitempty"`
	// +optional
	IPv6 *int `json:"ipv6,omitempty"`
}

// ServerParameters are the configurable fields of a Server.
type ServerParameters struct {
	// ServerType is the ID or name of the Server type this Server should be created with
	ServerType intstr.IntOrString `json:"serverType"`

	// Image is the ID or name of the Image the Server is created from
	Image intstr.IntOrString `json:"image"`

	// +optional
	SSHKeys *[]intstr.IntOrString `json:"sshKeys,omitempty"`

	// +optional
	Location *intstr.IntOrString `json:"location,omitempty"`

	// +optional
	Datacenter *intstr.IntOrString `json:"datacenter,omitempty"`

	// +optional
	UserData *string `json:"userData,omitempty"`

	// +optional
	StartAfterCreate *bool `json:"startAfterCreate,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// +optional
	Automount *bool `json:"automount,omitempty"`

	// +optional
	Volumes *[]int `json:"volumes,omitempty"`

	// +optional
	Networks *[]int `json:"networks,omitempty"`

	// +optional
	Firewalls *[]int `json:"firewalls,omitempty"`

	// +optional
	PlacementGroup *int `json:"placementGroup,omitempty"`

	// +optional
	PublicNet *PublicNetwork `json:"publicNet,omitempty"`
}

// ServerObservation are the observable fields of a Server.
type ServerObservation struct {
	Id      int          `json:"id"`
	Created *metav1.Time `json:"created,omitempty"`
	Status  string       `json:"status"`
	DNS     string       `json:"dns"`
	IPv4    string       `json:"ipv4"`
	IPv6    string       `json:"ipv6"`
}

// A ServerSpec defines the desired state of a Server.
type ServerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ServerParameters `json:"forProvider"`
}

// A ServerStatus represents the observed state of a Server.
type ServerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ServerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Server is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name",priority=1
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="DNS",type="string",JSONPath=".status.atProvider.dns",priority=10
// +kubebuilder:printcolumn:name="IPv4",type="string",JSONPath=".status.atProvider.ipv4"
// +kubebuilder:printcolumn:name="IPv6",type="string",JSONPath=".status.atProvider.ipv6",priority=10
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,hetzner}
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec"`
	Status ServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

// Server type metadata.
var (
	ServerKind             = reflect.TypeOf(Server{}).Name()
	ServerGroupKind        = schema.GroupKind{Group: Group, Kind: ServerKind}.String()
	ServerKindAPIVersion   = ServerKind + "." + SchemeGroupVersion.String()
	ServerGroupVersionKind = SchemeGroupVersion.WithKind(ServerKind)
)

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
