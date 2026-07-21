/*
Copyright 2024.

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
)

type AreaProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AreaSpec struct {
	Game           string         `json:"game"`
	World          string         `json:"world"`
	Description    string         `json:"description,omitempty"`
	ConnectedAreas []string       `json:"connectedAreas,omitempty"`
	Properties     []AreaProperty `json:"properties,omitempty"`
}

// AreaStatus defines the observed state of Area
type AreaStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="GAME",type=string,JSONPath=`.spec.game`
//+kubebuilder:printcolumn:name="WORLD",type=string,JSONPath=`.spec.world`

// Area is the Schema for the areas API
type Area struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AreaSpec   `json:"spec,omitempty"`
	Status AreaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AreaList contains a list of Area
type AreaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Area `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Area{}, &AreaList{})
}
