package v1alpha1

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestKindsAreRegistered(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	for _, kind := range []string{"MachineClass", "MachineClassList", "Machine", "MachineList", "MachineClaim", "MachineClaimList"} {
		gvk := SchemeGroupVersion.WithKind(kind)
		if _, err := scheme.New(gvk); err != nil {
			t.Errorf("scheme.New(%s): %v", gvk, err)
		}
	}
}

func TestDeepCopiesDoNotAliasMutableState(t *testing.T) {
	original := &MachineClass{
		Spec: MachineClassSpec{Definition: &MachineClassDefinition{
			Architecture: "amd64",
			Capabilities: []string{"console"},
			Capacity:     MachineCapacitySelector{MatchLabels: map[string]string{"tier": "metal"}},
		}},
	}
	copy := original.DeepCopy()
	copy.Spec.Definition.Capabilities[0] = "changed"
	copy.Spec.Definition.Capacity.MatchLabels["tier"] = "changed"
	if original.Spec.Definition.Capabilities[0] != "console" || original.Spec.Definition.Capacity.MatchLabels["tier"] != "metal" {
		t.Fatal("DeepCopy aliases MachineClass definition state")
	}
}
