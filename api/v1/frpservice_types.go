/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FrpServiceSpec defines the desired state of FrpService
type FrpServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:MinLength=0
	Address string `json:"address"`

	//+kubebuilder:validation:Minimum=0
	Port int32 `json:"port"`

	//+kubebuilder:validation:MinLength=0
	//+optional
	Token string `json:"token"`
}

// FrpServiceStatus defines the observed state of FrpService
type FrpServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FrpService is the Schema for the frpservices API
type FrpService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrpServiceSpec   `json:"spec,omitempty"`
	Status FrpServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrpServiceList contains a list of FrpService
type FrpServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrpService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrpService{}, &FrpServiceList{})
}
