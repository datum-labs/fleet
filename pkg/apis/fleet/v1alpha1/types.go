package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// LocalObjectReference identifies an object in the same Project control plane.
type LocalObjectReference struct {
	Name string `json:"name"`
}

// TypedObjectReference identifies a namespaced provider object.
type TypedObjectReference struct {
	APIGroup  string `json:"apiGroup"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// MachineClassSource identifies a class in another Project. Project is a
// routing address only; controllers must separately authorize its use.
type MachineClassSource struct {
	Project string `json:"project"`
	Name    string `json:"name"`
}

// MachineClassDefinition describes supply selected from a provider Project.
type MachineClassDefinition struct {
	Architecture string                  `json:"architecture"`
	Capabilities []string                `json:"capabilities,omitempty"`
	Capacity     MachineCapacitySelector `json:"capacity"`
}

type MachineCapacitySelector struct {
	Namespace        string               `json:"namespace"`
	MatchLabels      map[string]string    `json:"matchLabels,omitempty"`
	ProviderClassRef TypedObjectReference `json:"providerClassRef"`
}

type MachineClassSpec struct {
	Source     *MachineClassSource     `json:"source,omitempty"`
	Definition *MachineClassDefinition `json:"definition,omitempty"`
}

type MachineClassStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineClassSpec   `json:"spec"`
	Status            MachineClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClass `json:"items"`
}

type MachineSpec struct {
	ProviderRef TypedObjectReference `json:"providerRef"`
}

// MachineReservation is Fleet-managed exclusive reservation state. It is a
// pointer to the consumer binding, not a second binding or effects authority.
type MachineReservation struct {
	ConsumerProjectUID     string `json:"consumerProjectUID"`
	ConsumerProjectAddress string `json:"consumerProjectAddress"`
	Namespace              string `json:"namespace"`
	BindingName            string `json:"bindingName"`
	Epoch                  int64  `json:"epoch"`
}

type MachineStatus struct {
	ObservedGeneration int64               `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition  `json:"conditions,omitempty"`
	Reservation        *MachineReservation `json:"reservation,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineSpec   `json:"spec"`
	Status            MachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

type MachineClaimSpec struct {
	MachineClassRef LocalObjectReference `json:"machineClassRef"`
	LocationRef     LocalObjectReference `json:"locationRef"`
}

type MachineClaimPhase string

const (
	MachineClaimPhasePending   MachineClaimPhase = "Pending"
	MachineClaimPhaseBound     MachineClaimPhase = "Bound"
	MachineClaimPhaseReleasing MachineClaimPhase = "Releasing"
)

type MachineClaimStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Phase              MachineClaimPhase  `json:"phase,omitempty"`
	AllocationHandle   string             `json:"allocationHandle,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineClaimSpec   `json:"spec"`
	Status            MachineClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClaim `json:"items"`
}
