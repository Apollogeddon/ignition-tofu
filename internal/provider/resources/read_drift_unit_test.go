package resources

import (
	"context"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const smtpDescribedConfig = `
	provider "ignition" {
		host  = "http://mock-host"
		token = "mock-token"
	}
	resource "ignition_smtp_profile" "test" {
		name        = "smtp"
		description = "managed"
		hostname    = "smtp.test.com"
		port        = 25
	}
`

func TestUnitRead_DescriptionClearedOutsideOpenTofuIsDrift(t *testing.T) {
	description := ""
	mockClient := &client.MockClient{
		CreateSMTPProfileFunc: func(_ context.Context, item client.ResourceResponse[client.SMTPProfileConfig]) (*client.ResourceResponse[client.SMTPProfileConfig], error) {
			description = item.Description
			item.Signature = "sig"
			return &item, nil
		},
		UpdateSMTPProfileFunc: func(_ context.Context, item client.ResourceResponse[client.SMTPProfileConfig]) (*client.ResourceResponse[client.SMTPProfileConfig], error) {
			description = item.Description
			item.Signature = "sig"
			return &item, nil
		},
		GetSMTPProfileFunc: func(_ context.Context, name string) (*client.ResourceResponse[client.SMTPProfileConfig], error) {
			return &client.ResourceResponse[client.SMTPProfileConfig]{
				Name:        name,
				Enabled:     base.BoolPtr(true),
				Description: description,
				Signature:   "sig",
				Config: client.SMTPProfileConfig{
					Profile: client.SMTPProfileProfile{Type: "smtp.classic"},
					Settings: client.SMTPProfileSettings{
						Settings: &client.SMTPProfileSettingsClassic{Hostname: "smtp.test.com", Port: 25},
					},
				},
			}, nil
		},
		DeleteSMTPProfileFunc: func(context.Context, string, string) error { return nil },
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
				ResourceFactory: NewSMTPProfileResource,
				Client:          mockClient,
			}),
		},
		Steps: []resource.TestStep{
			{Config: smtpDescribedConfig},
			{
				PreConfig: func() { description = "" }, // cleared on the gateway
				Config:    smtpDescribedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("ignition_smtp_profile.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})

	if description != "managed" {
		t.Errorf("expected the apply to restore the description, got %q", description)
	}
}
