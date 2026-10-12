package v1alpha1

import (
	"errors"
	"fmt"
	"reflect"
)

func (s MachineClassSpec) Validate() error {
	if (s.Source == nil) == (s.Definition == nil) {
		return errors.New("exactly one of source and definition must be set")
	}
	if s.Source != nil && (s.Source.Project == "" || s.Source.Name == "") {
		return errors.New("source project and name are required")
	}
	if s.Definition != nil {
		if s.Definition.Arch == "" || s.Definition.Capacity.Namespace == "" {
			return errors.New("definition arch and explicit capacity namespace are required")
		}
		p := s.Definition.Capacity.ProviderClassRef
		if p.APIGroup == "" || p.Kind == "" || p.Namespace == "" || p.Name == "" {
			return errors.New("definition requires a fully typed, namespaced providerClassRef")
		}
	}
	return nil
}

func ValidateMachineClassUpdate(old, next MachineClass) error {
	if old.UID != next.UID {
		return errors.New("UID cannot change")
	}
	if old.Spec.Source != nil || next.Spec.Source != nil {
		if old.Spec.Source == nil || next.Spec.Source == nil || *old.Spec.Source != *next.Spec.Source {
			return errors.New("machine class mode and source are immutable")
		}
	}
	return next.Spec.Validate()
}

func (s MachineClaimSpec) Validate() error {
	if s.MachineClassRef.Name == "" || s.LocationRef.Name == "" {
		return errors.New("class and location are required")
	}
	if s.PowerState != PowerStateOn && s.PowerState != PowerStateOff {
		return errors.New("powerState must be On or Off")
	}
	return nil
}

func ValidateBoundClaimUpdate(old, next MachineClaim) error {
	if old.UID != next.UID {
		return errors.New("UID cannot change")
	}
	if old.Spec.MachineClassRef != next.Spec.MachineClassRef || old.Spec.LocationRef != next.Spec.LocationRef {
		return errors.New("class and placement are immutable after binding")
	}
	if !equalRef(old.Spec.InitialInstallationRef, next.Spec.InitialInstallationRef) {
		return errors.New("initialInstallationRef is immutable after binding")
	}
	if !reflect.DeepEqual(old.Spec.NetworkAttachmentRefs, next.Spec.NetworkAttachmentRefs) {
		return errors.New("networkAttachmentRefs are immutable after binding")
	}
	return next.Spec.Validate()
}

func (s InstallationSpec) Validate() error {
	if s.ClaimRef.Name == "" || s.ImageRef.Name == "" {
		return errors.New("claimRef and imageRef are required")
	}
	switch s.ExpectedPredecessor.Type {
	case PredecessorNone:
		if s.ExpectedPredecessor.Name != "" || s.ExpectedPredecessor.UID != "" {
			return errors.New("None predecessor cannot include name or UID")
		}
	case PredecessorInstallation:
		if s.ExpectedPredecessor.Name == "" || s.ExpectedPredecessor.UID == "" {
			return errors.New("Installation predecessor requires name and immutable UID")
		}
	default:
		return fmt.Errorf("unsupported predecessor type %q", s.ExpectedPredecessor.Type)
	}
	return nil
}

func ValidateInstallationUpdate(old, next Installation) error {
	if old.UID != next.UID {
		return errors.New("UID cannot change")
	}
	if !reflect.DeepEqual(old.Spec, next.Spec) {
		return errors.New("installation request is immutable")
	}
	return next.Spec.Validate()
}

func (b BindingRecord) CanRelease() error {
	if b.Phase != BindingReleasing {
		return errors.New("binding must be Releasing")
	}
	if len(b.ReleaseObligations) != 0 {
		return fmt.Errorf("release obligations remain: %v", b.ReleaseObligations)
	}
	if b.QuotaHoldID == "" {
		return errors.New("missing durable quota hold identity")
	}
	return nil
}

func (b BindingRecord) ValidateIdentity() error {
	if b.ID == "" || b.Epoch == 0 || b.ConsumerProjectUID == "" || b.ClaimUID == "" || b.LocalMachineClassUID == "" || b.ProviderProjectUID == "" || b.CanonicalMachineClassUID == "" || b.MachineUID == "" {
		return errors.New("binding identity is incomplete")
	}
	return nil
}

func equalRef(a, b *ObjectReference) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
