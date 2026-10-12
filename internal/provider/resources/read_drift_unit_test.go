package resources

import (
	"context"
	"testing"

	"github.com/apollogeddon/ignition-tofu/internal/client"
	"github.com/apollogeddon/ignition-tofu/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
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

const providerBlock = `
	provider "ignition" {
		host  = "http://mock-host"
		token = "mock-token"
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

// gateway stores what OpenTofu sends, so a test can change it as if from outside.
type gateway[T any] struct{ item client.ResourceResponse[T] }

func (g *gateway[T]) save(_ context.Context, item client.ResourceResponse[T]) (*client.ResourceResponse[T], error) {
	item.Signature = "sig"
	g.item = item
	return &item, nil
}

func (g *gateway[T]) get(context.Context, string) (*client.ResourceResponse[T], error) {
	item := g.item
	return &item, nil
}

func (g *gateway[T]) remove(context.Context, string, string) error { return nil }

// expectDrift applies config, clears a field on the gateway, and expects the
// next plan to restore it.
func expectDrift(t *testing.T, factory func() fwresource.Resource, mock *client.MockClient, config, address string, clear func()) {
	t.Helper()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{ResourceFactory: factory, Client: mock}),
		},
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: clear,
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
				},
			},
		},
	})
}

func TestUnitRead_FieldsClearedOutsideOpenTofuAreDrift(t *testing.T) {
	t.Run("database connection translator", func(t *testing.T) {
		gw := &gateway[client.DatabaseConfig]{}
		expectDrift(t, NewDatabaseConnectionResource, &client.MockClient{
			CreateDatabaseConnectionFunc: gw.save, UpdateDatabaseConnectionFunc: gw.save,
			GetDatabaseConnectionFunc: gw.get, DeleteDatabaseConnectionFunc: gw.remove,
		}, providerBlock+`
			resource "ignition_database_connection" "test" {
				name        = "db"
				type        = "MySQL"
				translator  = "MYSQL"
				connect_url = "jdbc:mysql://localhost"
			}
		`, "ignition_database_connection.test", func() { gw.item.Config.Translator = "" })
	})

	t.Run("user source failover profile", func(t *testing.T) {
		gw := &gateway[client.UserSourceConfig]{}
		expectDrift(t, NewUserSourceResource, &client.MockClient{
			CreateUserSourceFunc: gw.save, UpdateUserSourceFunc: gw.save,
			GetUserSourceFunc: gw.get, DeleteUserSourceFunc: gw.remove,
		}, providerBlock+`
			resource "ignition_user_source" "test" {
				name             = "users"
				type             = "INTERNAL"
				failover_profile = "backup"
				failover_mode    = "HARD"
			}
		`, "ignition_user_source.test", func() { gw.item.Config.Profile.FailoverProfile = "" })
	})

	t.Run("audit profile remote server", func(t *testing.T) {
		gw := &gateway[client.AuditProfileConfig]{}
		expectDrift(t, NewAuditProfileResource, &client.MockClient{
			CreateAuditProfileFunc: gw.save, UpdateAuditProfileFunc: gw.save,
			GetAuditProfileFunc: gw.get, DeleteAuditProfileFunc: gw.remove,
		}, providerBlock+`
			resource "ignition_audit_profile" "test" {
				name           = "audit"
				type           = "remote"
				remote_server  = "central"
				remote_profile = "audit"
			}
		`, "ignition_audit_profile.test", func() { gw.item.Config.Settings.RemoteServer = "" })
	})

	t.Run("alarm notification profile hostname", func(t *testing.T) {
		gw := &gateway[client.AlarmNotificationProfileConfig]{}
		expectDrift(t, NewAlarmNotificationProfileResource, &client.MockClient{
			CreateAlarmNotificationProfileFunc: gw.save, UpdateAlarmNotificationProfileFunc: gw.save,
			GetAlarmNotificationProfileFunc: gw.get, DeleteAlarmNotificationProfileFunc: gw.remove,
		}, providerBlock+`
			resource "ignition_alarm_notification_profile" "test" {
				name = "email"
				type = "EmailNotificationProfileType"
				email_config {
					use_smtp_profile = false
					hostname         = "smtp.test.com"
					port             = 25
					ssl_enabled      = false
				}
			}
		`, "ignition_alarm_notification_profile.test", func() {
			delete(gw.item.Config.Settings["settings"].(map[string]any), "hostname")
		})
	})
}

// An empty string set in config must survive the gateway's empty response,
// or apply fails with an inconsistent result.
func TestUnitRead_EmptyStringsSetInConfigStayEmpty(t *testing.T) {
	smtp := &gateway[client.SMTPProfileConfig]{}
	tags := &gateway[client.TagProviderConfig]{}
	db := &gateway[client.DatabaseConfig]{}
	var project client.Project
	mock := &client.MockClient{
		CreateSMTPProfileFunc: smtp.save, UpdateSMTPProfileFunc: smtp.save,
		GetSMTPProfileFunc: smtp.get, DeleteSMTPProfileFunc: smtp.remove,
		CreateTagProviderFunc: tags.save, UpdateTagProviderFunc: tags.save,
		GetTagProviderFunc: tags.get, DeleteTagProviderFunc: tags.remove,
		CreateDatabaseConnectionFunc: db.save, UpdateDatabaseConnectionFunc: db.save,
		GetDatabaseConnectionFunc: db.get, DeleteDatabaseConnectionFunc: db.remove,
		CreateProjectFunc: func(_ context.Context, p client.Project) (*client.Project, error) { project = p; return &p, nil },
		UpdateProjectFunc: func(_ context.Context, p client.Project) (*client.Project, error) { project = p; return &p, nil },
		GetProjectFunc:    func(context.Context, string) (*client.Project, error) { p := project; return &p, nil },
		DeleteProjectFunc: func(context.Context, string) error { return nil },
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"ignition": providerserver.NewProtocol6WithError(&base.TestProvider{
				Client: mock,
				ResourceFactories: []func() fwresource.Resource{
					NewSMTPProfileResource, NewTagProviderResource, NewDatabaseConnectionResource, NewProjectResource,
				},
			}),
		},
		Steps: []resource.TestStep{{
			Config: providerBlock + `
				resource "ignition_smtp_profile" "test" {
					name     = "smtp"
					hostname = "smtp.test.com"
					port     = 25
					username = ""
				}
				resource "ignition_tag_provider" "test" {
					name        = "tags"
					type        = "STANDARD"
					description = ""
				}
				resource "ignition_database_connection" "test" {
					name        = "db"
					type        = "MySQL"
					translator  = "MYSQL"
					connect_url = "jdbc:mysql://localhost"
					username    = ""
				}
				resource "ignition_project" "test" {
					name  = "project"
					title = ""
				}
			`,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
		}},
	})
}
