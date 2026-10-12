package v1alpha1

import "testing"

func TestMachineClassModes(t *testing.T) {
	ref := &SourceReference{Project: "datum-metal", Name: "standard"}
	for name, spec := range map[string]MachineClassSpec{
		"neither": {},
		"both":    {Source: ref, Definition: validDefinition()},
	} {
		t.Run(name, func(t *testing.T) {
			if spec.Validate() == nil {
				t.Fatal("expected rejection")
			}
		})
	}
	if err := (MachineClassSpec{Source: ref}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceRetargetRejected(t *testing.T) {
	old := MachineClass{UID: "local-class-uid", Spec: MachineClassSpec{Source: &SourceReference{Project: "p1", Name: "c"}}}
	next := old
	next.Spec.Source = &SourceReference{Project: "p2", Name: "c"}
	if ValidateMachineClassUpdate(old, next) == nil {
		t.Fatal("retargeting must require replacement")
	}
}

func TestInstallationRequiresPredecessorUID(t *testing.T) {
	s := InstallationSpec{ClaimRef: ObjectReference{"claim"}, ImageRef: ObjectReference{"image"}, ExpectedPredecessor: ExpectedPredecessor{Type: PredecessorInstallation, Name: "old"}}
	if s.Validate() == nil {
		t.Fatal("name reuse must not satisfy predecessor fencing")
	}
	s.ExpectedPredecessor.UID = "old-uid"
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBoundClaimAllowsPowerIntentButNotRetarget(t *testing.T) {
	old := MachineClaim{UID: "claim-uid", Spec: MachineClaimSpec{MachineClassRef: ObjectReference{"class"}, LocationRef: ObjectReference{"loc"}, PowerState: PowerStateOff, NetworkAttachmentRefs: []ObjectReference{{"network"}}}}
	next := old
	next.Spec.PowerState = PowerStateOn
	if err := ValidateBoundClaimUpdate(old, next); err != nil {
		t.Fatal(err)
	}
	next.Spec.MachineClassRef.Name = "other"
	if ValidateBoundClaimUpdate(old, next) == nil {
		t.Fatal("bound class change must fail")
	}
	next = old
	next.Spec.NetworkAttachmentRefs = []ObjectReference{{"other-network"}}
	if ValidateBoundClaimUpdate(old, next) == nil {
		t.Fatal("bound network attachment change must fail")
	}
}

func TestInstallationUpdateUsesSemanticEquality(t *testing.T) {
	a := &ObjectReference{Name: "bootstrap"}
	b := &ObjectReference{Name: "bootstrap"}
	old := Installation{UID: "install-uid", Spec: InstallationSpec{ClaimRef: ObjectReference{"claim"}, ExpectedPredecessor: ExpectedPredecessor{Type: PredecessorNone}, ImageRef: ObjectReference{"image"}, BootstrapConfigRef: a}}
	next := old
	next.Spec.BootstrapConfigRef = b
	if err := ValidateInstallationUpdate(old, next); err != nil {
		t.Fatalf("equal values at different addresses must be accepted: %v", err)
	}
}

func TestReleaseRetainsReservationUntilObligationsClear(t *testing.T) {
	b := BindingRecord{Phase: BindingReleasing, QuotaHoldID: "quota-1", ReleaseObligations: []string{"network-detach", "disk-clean"}}
	if b.CanRelease() == nil {
		t.Fatal("unclean asset must remain reserved")
	}
	b.ReleaseObligations = nil
	if err := b.CanRelease(); err != nil {
		t.Fatal(err)
	}
}

func TestBindingIdentityUsesUIDsAcrossOrWithinProjects(t *testing.T) {
	b := BindingRecord{ID: "binding-1", Epoch: 1, ConsumerProjectUID: "consumer-uid", ClaimUID: "claim-uid", LocalMachineClassUID: "local-class-uid", ProviderProjectUID: "provider-uid", CanonicalMachineClassUID: "source-class-uid", MachineUID: "machine-uid"}
	if err := b.ValidateIdentity(); err != nil {
		t.Fatal(err)
	}
	b.ProviderProjectUID = b.ConsumerProjectUID
	if err := b.ValidateIdentity(); err != nil {
		t.Fatalf("definitions may serve claims in their own project: %v", err)
	}
}

func validDefinition() *MachineClassDefinition {
	return &MachineClassDefinition{Arch: "amd64", Capacity: MachineCapacitySelector{Namespace: "infra", ProviderClassRef: TypedObjectReference{APIGroup: "baremetal.fleet.datumapis.com", Kind: "BareMetalMachineClass", Namespace: "infra", Name: "qualified"}}}
}
