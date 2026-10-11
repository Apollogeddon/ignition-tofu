package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// The registry docs are generated from these descriptions, so an attribute without one shows up there blank.
func TestUnitSchemaDescriptions(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		t.Fatalf("%s: %s", d.Summary, d.Detail)
	}

	checkBlock(t, "provider", resp.Provider.Block)
	for name, s := range resp.ResourceSchemas {
		checkBlock(t, name, s.Block)
	}
	for name, s := range resp.DataSourceSchemas {
		checkBlock(t, "data."+name, s.Block)
	}
}

func checkBlock(t *testing.T, where string, block *tfprotov6.SchemaBlock) {
	t.Helper()
	if block.Description == "" {
		t.Errorf("%s has no description", where)
	}
	checkAttributes(t, where, block.Attributes)
	for _, nested := range block.BlockTypes {
		checkBlock(t, where+"."+nested.TypeName, nested.Block)
	}
}

func checkAttributes(t *testing.T, where string, attributes []*tfprotov6.SchemaAttribute) {
	t.Helper()
	for _, a := range attributes {
		// tfplugindocs describes id itself
		if a.Description == "" && a.Name != "id" {
			t.Errorf("%s.%s has no description", where, a.Name)
		}
		if a.NestedType != nil {
			checkAttributes(t, where+"."+a.Name, a.NestedType.Attributes)
		}
	}
}
