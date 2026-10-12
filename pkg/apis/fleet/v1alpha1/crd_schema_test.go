package v1alpha1

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

type crdDocument struct {
	Spec struct {
		Group string `json:"group"`
		Scope string `json:"scope"`
		Names struct {
			Kind string `json:"kind"`
		} `json:"names"`
		Versions []struct {
			Name         string                 `json:"name"`
			Served       bool                   `json:"served"`
			Storage      bool                   `json:"storage"`
			Subresources map[string]interface{} `json:"subresources"`
			Schema       struct {
				OpenAPIV3Schema schemaNode `json:"openAPIV3Schema"`
			} `json:"schema"`
		} `json:"versions"`
	} `json:"spec"`
}

type schemaNode struct {
	Type                  string                `json:"type"`
	Properties            map[string]schemaNode `json:"properties"`
	Items                 *schemaNode           `json:"items"`
	AdditionalProperties  *schemaNode           `json:"additionalProperties"`
	PreserveUnknownFields bool                  `json:"x-kubernetes-preserve-unknown-fields"`
	Validations           []schemaValidation    `json:"x-kubernetes-validations"`
}

type schemaValidation struct {
	Rule string `json:"rule"`
}

func TestCRDSchemas(t *testing.T) {
	tests := []struct {
		file, kind, scope string
		rules             []string
	}{
		{"fleet.datumapis.com_machineclasses.yaml", "MachineClass", "Cluster", []string{"has(self.source) != has(self.definition)", "self == oldSelf"}},
		{"fleet.datumapis.com_machines.yaml", "Machine", "Namespaced", []string{"self == oldSelf"}},
		{"fleet.datumapis.com_machineclaims.yaml", "MachineClaim", "Namespaced", []string{"self.machineClassRef == oldSelf.machineClassRef", "self.locationRef == oldSelf.locationRef"}},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "..", "config", "crd", "bases", tt.file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var crd crdDocument
			if err := yaml.Unmarshal(data, &crd); err != nil {
				t.Fatalf("parse CRD: %v", err)
			}
			if crd.Spec.Group != GroupName || crd.Spec.Scope != tt.scope || crd.Spec.Names.Kind != tt.kind {
				t.Fatalf("unexpected identity: group=%q scope=%q kind=%q", crd.Spec.Group, crd.Spec.Scope, crd.Spec.Names.Kind)
			}
			if len(crd.Spec.Versions) != 1 || crd.Spec.Versions[0].Name != "v1alpha1" || !crd.Spec.Versions[0].Served || !crd.Spec.Versions[0].Storage {
				t.Fatalf("expected one served storage v1alpha1 version: %#v", crd.Spec.Versions)
			}
			version := crd.Spec.Versions[0]
			if _, ok := version.Subresources["status"]; !ok {
				t.Fatal("status subresource is required")
			}
			assertStructural(t, "root", version.Schema.OpenAPIV3Schema)
			spec := version.Schema.OpenAPIV3Schema.Properties["spec"]
			for _, want := range tt.rules {
				found := false
				for _, validation := range spec.Validations {
					found = found || validation.Rule == want
				}
				if !found {
					t.Errorf("missing CEL rule %q", want)
				}
			}
		})
	}
}

func assertStructural(t *testing.T, path string, node schemaNode) {
	t.Helper()
	if node.Type == "" {
		t.Errorf("%s has no type", path)
	}
	if node.PreserveUnknownFields {
		t.Errorf("%s preserves unknown fields", path)
	}
	for name, child := range node.Properties {
		assertStructural(t, path+"."+name, child)
	}
	if node.Items != nil {
		assertStructural(t, path+"[]", *node.Items)
	}
	if node.AdditionalProperties != nil {
		assertStructural(t, path+"{}", *node.AdditionalProperties)
	}
}
