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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SSHKeyParameters are the configurable fields of an SSHKey.
type SSHKeyParameters struct {
	PublicKey string `json:"publicKey"`

	// +optional
	Labels map[string]string `json:"labels"`
}

// SSHKeyObservation are the observable fields of a SSHKey.
type SSHKeyObservation struct {
	Id          int          `json:"id"`
	Created     *metav1.Time `json:"created,omitempty"`
	Fingerprint string       `json:"fingerprint"`
}

// A SSHKeySpec defines the desired state of a SSHKey.
type SSHKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SSHKeyParameters `json:"forProvider"`
}

// A SSHKeyStatus represents the observed state of a SSHKey.
type SSHKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SSHKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SSHKey is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,template}
type SSHKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHKeySpec   `json:"spec"`
	Status SSHKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SSHKeyList contains a list of SSHKey
type SSHKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHKey `json:"items"`
}

// SSHKey type metadata.
var (
	SSHKeyKind             = reflect.TypeOf(SSHKey{}).Name()
	SSHKeyGroupKind        = schema.GroupKind{Group: Group, Kind: SSHKeyKind}.String()
	SSHKeyKindAPIVersion   = SSHKeyKind + "." + SchemeGroupVersion.String()
	SSHKeyGroupVersionKind = SchemeGroupVersion.WithKind(SSHKeyKind)
)

func init() {
	SchemeBuilder.Register(&SSHKey{}, &SSHKeyList{})
}
