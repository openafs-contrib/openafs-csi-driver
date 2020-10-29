package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type OpenAFSProvSpec struct {
	ProvisionerName			string  `json:"provisionerName"`
	ProvisionerNameSpace		string	`json:"provisionerNameSpace"`
	ProvisionerImageName		string	`json:"provisionerImageName"`
}

type OpenAFSAttacherSpec struct {
	AttacherName			string  `json:"attacherName"`
	AttacherNameSpace		string	`json:"attacherNameSpace"`
	AttacherImageName		string	`json:"attacherImageName"`
}

type OpenAFSPluginSpec struct {
	PluginName			string  `json:"pluginName"`
	PluginNameSpace			string	`json:"pluginNameSpace"`
	DriverRegistrarImage		string	`json:"driverRegistrarImage"`
	PluginImage			string  `json:"pluginImage"`
	LivenessProbeImage		string 	`json:"livenessProbeImage"`
        AfsMount			string	`json:"afsMount"`
        Configmap			string  `json:"configmap"`
}

// OpenafsCSIAppSpec defines the desired state of OpenafsCSIApp
type OpenafsCSIAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ProvisionerSpec		OpenAFSProvSpec 	`json:"provisionerSpec"`
	AttacherSpec		OpenAFSAttacherSpec 	`json:"attacherSpec"`
	PluginSpec		OpenAFSPluginSpec	`json:"pluginSpec"`	
}

// OpenafsCSIAppStatus defines the observed state of OpenafsCSIApp
type OpenafsCSIAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenafsCSIApp is the Schema for the openafscsiapps API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=openafscsiapps,scope=Namespaced
type OpenafsCSIApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenafsCSIAppSpec   `json:"spec,omitempty"`
	Status OpenafsCSIAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenafsCSIAppList contains a list of OpenafsCSIApp
type OpenafsCSIAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenafsCSIApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenafsCSIApp{}, &OpenafsCSIAppList{})
}
