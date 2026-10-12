package v1alpha1

import "time"

type UID string
type ProjectUID string

// ObjectReference is a local, name-based request reference. Controllers resolve
// it to a UID before committing authority or causing external effects.
type ObjectReference struct {
	Name string `json:"name"`
}

type TypedObjectReference struct {
	APIGroup  string `json:"apiGroup"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type SourceReference struct {
	// Project is an address used for resolution. It is not authorization.
	Project string `json:"project"`
	Name    string `json:"name"`
}

type MachineClass struct {
	UID  UID              `json:"uid"`
	Spec MachineClassSpec `json:"spec"`
}

type MachineClassSpec struct {
	// Exactly one of Source and Definition must be set.
	Source     *SourceReference        `json:"source,omitempty"`
	Definition *MachineClassDefinition `json:"definition,omitempty"`
}

type MachineClassDefinition struct {
	Arch         string                  `json:"arch"`
	Capabilities []string                `json:"capabilities,omitempty"`
	Access       AccessContract          `json:"access"`
	Capacity     MachineCapacitySelector `json:"capacity"`
}

type AccessContract struct {
	ConsoleModes []string `json:"consoleModes,omitempty"`
	PowerControl bool     `json:"powerControl"`
}

type MachineCapacitySelector struct {
	Namespace        string               `json:"namespace"`
	MatchLabels      map[string]string    `json:"matchLabels,omitempty"`
	ProviderClassRef TypedObjectReference `json:"providerClassRef"`
}

type MachineClaim struct {
	UID  UID              `json:"uid"`
	Spec MachineClaimSpec `json:"spec"`
}

type MachineClaimSpec struct {
	MachineClassRef        ObjectReference   `json:"machineClassRef"`
	LocationRef            ObjectReference   `json:"locationRef"`
	PowerState             PowerState        `json:"powerState"`
	NetworkAttachmentRefs  []ObjectReference `json:"networkAttachmentRefs,omitempty"`
	InitialInstallationRef *ObjectReference  `json:"initialInstallationRef,omitempty"`
}

type PowerState string

const (
	PowerStateOff PowerState = "Off"
	PowerStateOn  PowerState = "On"
)

type Installation struct {
	UID  UID              `json:"uid"`
	Spec InstallationSpec `json:"spec"`
}

type InstallationSpec struct {
	ClaimRef            ObjectReference     `json:"claimRef"`
	ExpectedPredecessor ExpectedPredecessor `json:"expectedPredecessor"`
	ImageRef            ObjectReference     `json:"imageRef"`
	BootstrapConfigRef  *ObjectReference    `json:"bootstrapConfigRef,omitempty"`
}

type ExpectedPredecessor struct {
	Type PredecessorType `json:"type"`
	Name string          `json:"name,omitempty"`
	UID  UID             `json:"uid,omitempty"`
}

type PredecessorType string

const (
	PredecessorNone         PredecessorType = "None"
	PredecessorInstallation PredecessorType = "Installation"
)

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message,omitempty"`
	ObservedGeneration int64     `json:"observedGeneration"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	// These correlation fields prevent a stale observation from describing a new allocation.
	BindingID    string `json:"bindingID,omitempty"`
	BindingEpoch uint64 `json:"bindingEpoch,omitempty"`
	InputDigest  string `json:"inputDigest,omitempty"`
}

// BindingRecord is protected service state, deliberately separate from public
// MachineClaim status and retained outside the consumer project's deletion domain.
type BindingRecord struct {
	ID                        string       `json:"id"`
	Epoch                     uint64       `json:"epoch"`
	ConsumerProjectUID        ProjectUID   `json:"consumerProjectUID"`
	ClaimUID                  UID          `json:"claimUID"`
	LocalMachineClassUID      UID          `json:"localMachineClassUID"`
	ProviderProjectUID        ProjectUID   `json:"providerProjectUID"`
	CanonicalMachineClassUID  UID          `json:"canonicalMachineClassUID"`
	CanonicalClassDigest      string       `json:"canonicalClassDigest"`
	MachineUID                UID          `json:"machineUID"`
	AuthorizationPolicyDigest string       `json:"authorizationPolicyDigest"`
	QuotaHoldID               string       `json:"quotaHoldID"`
	Phase                     BindingPhase `json:"phase"`
	ReleaseObligations        []string     `json:"releaseObligations,omitempty"`
}

type BindingPhase string

const (
	BindingActive      BindingPhase = "Active"
	BindingReleasing   BindingPhase = "Releasing"
	BindingQuarantined BindingPhase = "Quarantined"
	BindingReleased    BindingPhase = "Released"
)

// InstallationPlan is protected, immutable execution authority persisted before effects.
type InstallationPlan struct {
	InstallationUID    UID        `json:"installationUID"`
	ConsumerProjectUID ProjectUID `json:"consumerProjectUID"`
	ClaimUID           UID        `json:"claimUID"`
	MachineUID         UID        `json:"machineUID"`
	BindingID          string     `json:"bindingID"`
	BindingEpoch       uint64     `json:"bindingEpoch"`
	PredecessorUID     UID        `json:"predecessorUID,omitempty"`
	ImageUID           UID        `json:"imageUID"`
	ImageDigest        string     `json:"imageDigest"`
	BootstrapUID       UID        `json:"bootstrapUID,omitempty"`
	InputsDigest       string     `json:"inputsDigest"`
	Executor           string     `json:"executor"`
	OperationID        string     `json:"operationID"`
}
