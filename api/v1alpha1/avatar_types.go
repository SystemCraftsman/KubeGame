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

type AttributeType struct {
	Name      string `json:"name"`
	ValueType string `json:"valueType"`
}

type InventoryType struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type AchievementType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CustomizationType struct {
	Name    string   `json:"name"`
	Options []string `json:"options"`
}

type AvatarSpec struct {
	Game               string              `json:"game"`
	Type               string              `json:"type"`
	Description        string              `json:"description,omitempty"`
	AttributeTypes     []AttributeType     `json:"attributeTypes,omitempty"`
	InventoryTypes     []InventoryType     `json:"inventoryTypes,omitempty"`
	AchievementTypes   []AchievementType   `json:"achievementTypes,omitempty"`
	CustomizationTypes []CustomizationType `json:"customizationTypes,omitempty"`
}

// AvatarStatus defines the observed state of Avatar
type AvatarStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Avatar is the Schema for the avatars API
type Avatar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AvatarSpec   `json:"spec,omitempty"`
	Status AvatarStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AvatarList contains a list of Avatar
type AvatarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Avatar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Avatar{}, &AvatarList{})
}
