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

type ItemEffect struct {
	Attribute string `json:"attribute"`
	Modifier  string `json:"modifier"`
}

type ItemSpec struct {
	Name      string       `json:"name"`
	Category  string       `json:"category"`
	Rarity    string       `json:"rarity,omitempty"`
	Stackable bool         `json:"stackable,omitempty"`
	MaxStack  int          `json:"maxStack,omitempty"`
	Duration  int          `json:"duration,omitempty"`
	Effects   []ItemEffect `json:"effects,omitempty"`
}

type ItemCatalogSpec struct {
	Game  string     `json:"game"`
	Items []ItemSpec `json:"items"`
}

type ItemCatalogStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type ItemCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ItemCatalogSpec   `json:"spec,omitempty"`
	Status ItemCatalogStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type ItemCatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ItemCatalog `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ItemCatalog{}, &ItemCatalogList{})
}
