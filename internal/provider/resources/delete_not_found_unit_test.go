package resources

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// smtpDeleteMock is smtpMock whose delete returns deleteErr, as a gateway does
// when the profile went between OpenTofu's refresh and its delete.
func smtpDeleteMock(deleteErr *error, deletes *int) func() (tfprotov6.ProviderServer, error) {
	exists, creates := false, 0
	var getErr error
	mockClient := smtpMock(&exists, &creates, &getErr)
	mockClient.DeleteSMTPProfileFunc = func(context.Context, string, string) error {
		*deletes++
		if *deleteErr != nil {
			return *deleteErr
		}
		exists = false
		return nil
	}
	return providerserver.NewProtocol6WithError(&base.TestProvider{
		ResourceFactory: NewSMTPProfileResource,
		Client:          mockClient,
	})
}

func TestUnitDelete_AlreadyDeletedIsSuccess(t *testing.T) {
	deleteErr, deletes := errGatewayNotFound, 0

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"ignition": smtpDeleteMock(&deleteErr, &deletes),
		},
		Steps: []resource.TestStep{
			{Config: smtpDeletedConfig},
			{Config: smtpDeletedConfig, Destroy: true},
		},
	})

	if deletes == 0 {
		t.Error("expected destroy to call delete")
	}
}

func TestUnitDelete_OtherErrorsStillFail(t *testing.T) {
	var deleteErr error
	deletes := 0

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"ignition": smtpDeleteMock(&deleteErr, &deletes),
		},
		Steps: []resource.TestStep{
			{Config: smtpDeletedConfig},
			{
				PreConfig:   func() { deleteErr = fmt.Errorf("status: 500, body: boom") },
				Config:      smtpDeletedConfig,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Error deleting resource`),
			},
			{
				// let the post-test destroy succeed
				PreConfig: func() { deleteErr = nil },
				Config:    smtpDeletedConfig,
			},
		},
	})
}
