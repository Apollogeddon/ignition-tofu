package client

import "context"

type MockClient struct {
	GetResourceFunc                    func(ctx context.Context, rt, n string, d any) error
	CreateResourceFunc                 func(ctx context.Context, rt string, i, d any) error
	UpdateResourceFunc                 func(ctx context.Context, rt string, i, d any) error
	DeleteResourceFunc                 func(ctx context.Context, rt, n, s string) error
	GetResourceWithModuleFunc          func(ctx context.Context, m, rt, n string, d any) error
	CreateResourceWithModuleFunc       func(ctx context.Context, m, rt string, i, d any) error
	UpdateResourceWithModuleFunc       func(ctx context.Context, m, rt string, i, d any) error
	DeleteResourceWithModuleFunc       func(ctx context.Context, m, rt, n, s string) error
	EncryptSecretFunc                  func(ctx context.Context, p string) (*IgnitionSecret, error)
	GetProjectFunc                     func(ctx context.Context, n string) (*Project, error)
	CreateProjectFunc                  func(ctx context.Context, p Project) (*Project, error)
	UpdateProjectFunc                  func(ctx context.Context, p Project) (*Project, error)
	DeleteProjectFunc                  func(ctx context.Context, n string) error
	GetDatabaseConnectionFunc          func(ctx context.Context, n string) (*ResourceResponse[DatabaseConfig], error)
	CreateDatabaseConnectionFunc       func(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error)
	UpdateDatabaseConnectionFunc       func(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error)
	DeleteDatabaseConnectionFunc       func(ctx context.Context, n, s string) error
	GetUserSourceFunc                  func(ctx context.Context, n string) (*ResourceResponse[UserSourceConfig], error)
	CreateUserSourceFunc               func(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error)
	UpdateUserSourceFunc               func(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error)
	DeleteUserSourceFunc               func(ctx context.Context, n, s string) error
	GetTagProviderFunc                 func(ctx context.Context, n string) (*ResourceResponse[TagProviderConfig], error)
	CreateTagProviderFunc              func(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error)
	UpdateTagProviderFunc              func(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error)
	DeleteTagProviderFunc              func(ctx context.Context, n, s string) error
	GetAuditProfileFunc                func(ctx context.Context, n string) (*ResourceResponse[AuditProfileConfig], error)
	CreateAuditProfileFunc             func(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error)
	UpdateAuditProfileFunc             func(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error)
	DeleteAuditProfileFunc             func(ctx context.Context, n, s string) error
	GetAlarmNotificationProfileFunc    func(ctx context.Context, n string) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	CreateAlarmNotificationProfileFunc func(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	UpdateAlarmNotificationProfileFunc func(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	DeleteAlarmNotificationProfileFunc func(ctx context.Context, n, s string) error
	GetOpcUaConnectionFunc             func(ctx context.Context, n string) (*ResourceResponse[OpcUaConnectionConfig], error)
	CreateOpcUaConnectionFunc          func(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error)
	UpdateOpcUaConnectionFunc          func(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error)
	DeleteOpcUaConnectionFunc          func(ctx context.Context, n, s string) error
	GetAlarmJournalFunc                func(ctx context.Context, n string) (*ResourceResponse[AlarmJournalConfig], error)
	CreateAlarmJournalFunc             func(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error)
	UpdateAlarmJournalFunc             func(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error)
	DeleteAlarmJournalFunc             func(ctx context.Context, n, s string) error
	GetHistorianProviderFunc           func(ctx context.Context, n string) (*ResourceResponse[HistorianProviderConfig], error)
	CreateHistorianProviderFunc        func(ctx context.Context, i ResourceResponse[HistorianProviderConfig]) (*ResourceResponse[HistorianProviderConfig], error)
	UpdateHistorianProviderFunc        func(ctx context.Context, i ResourceResponse[HistorianProviderConfig]) (*ResourceResponse[HistorianProviderConfig], error)
	DeleteHistorianProviderFunc        func(ctx context.Context, n, s string) error
	GetSMTPProfileFunc                 func(ctx context.Context, n string) (*ResourceResponse[SMTPProfileConfig], error)
	CreateSMTPProfileFunc              func(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error)
	UpdateSMTPProfileFunc              func(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error)
	DeleteSMTPProfileFunc              func(ctx context.Context, n, s string) error
	GetStoreAndForwardFunc             func(ctx context.Context, n string) (*ResourceResponse[StoreAndForwardConfig], error)
	CreateStoreAndForwardFunc          func(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error)
	UpdateStoreAndForwardFunc          func(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error)
	DeleteStoreAndForwardFunc          func(ctx context.Context, n, s string) error
	GetIdentityProviderFunc            func(ctx context.Context, n string) (*ResourceResponse[IdentityProviderConfig], error)
	CreateIdentityProviderFunc         func(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error)
	UpdateIdentityProviderFunc         func(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error)
	DeleteIdentityProviderFunc         func(ctx context.Context, n, s string) error
	GetGanOutgoingFunc                 func(ctx context.Context, n string) (*ResourceResponse[GanOutgoingConfig], error)
	CreateGanOutgoingFunc              func(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error)
	UpdateGanOutgoingFunc              func(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error)
	DeleteGanOutgoingFunc              func(ctx context.Context, n, s string) error
	GetRedundancyConfigFunc            func(ctx context.Context) (*RedundancyConfig, error)
	UpdateRedundancyConfigFunc         func(ctx context.Context, c RedundancyConfig) error
	GetGanGeneralSettingsFunc          func(ctx context.Context) (*ResourceResponse[GanGeneralSettingsConfig], error)
	UpdateGanGeneralSettingsFunc       func(ctx context.Context, i ResourceResponse[GanGeneralSettingsConfig]) (*ResourceResponse[GanGeneralSettingsConfig], error)
	GetDeviceFunc                      func(ctx context.Context, n string) (*ResourceResponse[DeviceConfig], error)
	CreateDeviceFunc                   func(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error)
	UpdateDeviceFunc                   func(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error)
	DeleteDeviceFunc                   func(ctx context.Context, n, s string) error
}

func (m *MockClient) GetResource(ctx context.Context, rt, n string, d any) error {
	if m.GetResourceFunc != nil {
		return m.GetResourceFunc(ctx, rt, n, d)
	}
	return nil
}

func (m *MockClient) CreateResource(ctx context.Context, rt string, i, d any) error {
	if m.CreateResourceFunc != nil {
		return m.CreateResourceFunc(ctx, rt, i, d)
	}
	return nil
}

func (m *MockClient) UpdateResource(ctx context.Context, rt string, i, d any) error {
	if m.UpdateResourceFunc != nil {
		return m.UpdateResourceFunc(ctx, rt, i, d)
	}
	return nil
}

func (m *MockClient) DeleteResource(ctx context.Context, rt, n, s string) error {
	if m.DeleteResourceFunc != nil {
		return m.DeleteResourceFunc(ctx, rt, n, s)
	}
	return nil
}

func (m *MockClient) GetResourceWithModule(ctx context.Context, mod, rt, n string, d any) error {
	if m.GetResourceWithModuleFunc != nil {
		return m.GetResourceWithModuleFunc(ctx, mod, rt, n, d)
	}
	return nil
}

func (m *MockClient) CreateResourceWithModule(ctx context.Context, mod, rt string, i, d any) error {
	if m.CreateResourceWithModuleFunc != nil {
		return m.CreateResourceWithModuleFunc(ctx, mod, rt, i, d)
	}
	return nil
}

func (m *MockClient) UpdateResourceWithModule(ctx context.Context, mod, rt string, i, d any) error {
	if m.UpdateResourceWithModuleFunc != nil {
		return m.UpdateResourceWithModuleFunc(ctx, mod, rt, i, d)
	}
	return nil
}

func (m *MockClient) DeleteResourceWithModule(ctx context.Context, mod, rt, n, s string) error {
	if m.DeleteResourceWithModuleFunc != nil {
		return m.DeleteResourceWithModuleFunc(ctx, mod, rt, n, s)
	}
	return nil
}

func (m *MockClient) EncryptSecret(ctx context.Context, p string) (*IgnitionSecret, error) {
	if m.EncryptSecretFunc != nil {
		return m.EncryptSecretFunc(ctx, p)
	}
	return &IgnitionSecret{Type: "Embedded", Data: map[string]any{"value": p}}, nil
}

func (m *MockClient) GetProject(ctx context.Context, n string) (*Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, n)
	}
	return &Project{}, nil
}

func (m *MockClient) CreateProject(ctx context.Context, p Project) (*Project, error) {
	if m.CreateProjectFunc != nil {
		return m.CreateProjectFunc(ctx, p)
	}
	return &Project{}, nil
}

func (m *MockClient) UpdateProject(ctx context.Context, p Project) (*Project, error) {
	if m.UpdateProjectFunc != nil {
		return m.UpdateProjectFunc(ctx, p)
	}
	return &Project{}, nil
}

func (m *MockClient) DeleteProject(ctx context.Context, n string) error {
	if m.DeleteProjectFunc != nil {
		return m.DeleteProjectFunc(ctx, n)
	}
	return nil
}

func (m *MockClient) GetDatabaseConnection(ctx context.Context, n string) (*ResourceResponse[DatabaseConfig], error) {
	if m.GetDatabaseConnectionFunc != nil {
		return m.GetDatabaseConnectionFunc(ctx, n)
	}
	return &ResourceResponse[DatabaseConfig]{}, nil
}

func (m *MockClient) CreateDatabaseConnection(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error) {
	if m.CreateDatabaseConnectionFunc != nil {
		return m.CreateDatabaseConnectionFunc(ctx, i)
	}
	return &ResourceResponse[DatabaseConfig]{}, nil
}

func (m *MockClient) UpdateDatabaseConnection(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error) {
	if m.UpdateDatabaseConnectionFunc != nil {
		return m.UpdateDatabaseConnectionFunc(ctx, i)
	}
	return &ResourceResponse[DatabaseConfig]{}, nil
}

func (m *MockClient) DeleteDatabaseConnection(ctx context.Context, n, s string) error {
	if m.DeleteDatabaseConnectionFunc != nil {
		return m.DeleteDatabaseConnectionFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetUserSource(ctx context.Context, n string) (*ResourceResponse[UserSourceConfig], error) {
	if m.GetUserSourceFunc != nil {
		return m.GetUserSourceFunc(ctx, n)
	}
	return &ResourceResponse[UserSourceConfig]{}, nil
}

func (m *MockClient) CreateUserSource(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error) {
	if m.CreateUserSourceFunc != nil {
		return m.CreateUserSourceFunc(ctx, i)
	}
	return &ResourceResponse[UserSourceConfig]{}, nil
}

func (m *MockClient) UpdateUserSource(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error) {
	if m.UpdateUserSourceFunc != nil {
		return m.UpdateUserSourceFunc(ctx, i)
	}
	return &ResourceResponse[UserSourceConfig]{}, nil
}

func (m *MockClient) DeleteUserSource(ctx context.Context, n, s string) error {
	if m.DeleteUserSourceFunc != nil {
		return m.DeleteUserSourceFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetTagProvider(ctx context.Context, n string) (*ResourceResponse[TagProviderConfig], error) {
	if m.GetTagProviderFunc != nil {
		return m.GetTagProviderFunc(ctx, n)
	}
	return &ResourceResponse[TagProviderConfig]{}, nil
}

func (m *MockClient) CreateTagProvider(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error) {
	if m.CreateTagProviderFunc != nil {
		return m.CreateTagProviderFunc(ctx, i)
	}
	return &ResourceResponse[TagProviderConfig]{}, nil
}

func (m *MockClient) UpdateTagProvider(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error) {
	if m.UpdateTagProviderFunc != nil {
		return m.UpdateTagProviderFunc(ctx, i)
	}
	return &ResourceResponse[TagProviderConfig]{}, nil
}

func (m *MockClient) DeleteTagProvider(ctx context.Context, n, s string) error {
	if m.DeleteTagProviderFunc != nil {
		return m.DeleteTagProviderFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetAuditProfile(ctx context.Context, n string) (*ResourceResponse[AuditProfileConfig], error) {
	if m.GetAuditProfileFunc != nil {
		return m.GetAuditProfileFunc(ctx, n)
	}
	return &ResourceResponse[AuditProfileConfig]{}, nil
}

func (m *MockClient) CreateAuditProfile(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error) {
	if m.CreateAuditProfileFunc != nil {
		return m.CreateAuditProfileFunc(ctx, i)
	}
	return &ResourceResponse[AuditProfileConfig]{}, nil
}

func (m *MockClient) UpdateAuditProfile(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error) {
	if m.UpdateAuditProfileFunc != nil {
		return m.UpdateAuditProfileFunc(ctx, i)
	}
	return &ResourceResponse[AuditProfileConfig]{}, nil
}

func (m *MockClient) DeleteAuditProfile(ctx context.Context, n, s string) error {
	if m.DeleteAuditProfileFunc != nil {
		return m.DeleteAuditProfileFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetAlarmNotificationProfile(ctx context.Context, n string) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	if m.GetAlarmNotificationProfileFunc != nil {
		return m.GetAlarmNotificationProfileFunc(ctx, n)
	}
	return &ResourceResponse[AlarmNotificationProfileConfig]{}, nil
}

func (m *MockClient) CreateAlarmNotificationProfile(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	if m.CreateAlarmNotificationProfileFunc != nil {
		return m.CreateAlarmNotificationProfileFunc(ctx, i)
	}
	return &ResourceResponse[AlarmNotificationProfileConfig]{}, nil
}

func (m *MockClient) UpdateAlarmNotificationProfile(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	if m.UpdateAlarmNotificationProfileFunc != nil {
		return m.UpdateAlarmNotificationProfileFunc(ctx, i)
	}
	return &ResourceResponse[AlarmNotificationProfileConfig]{}, nil
}

func (m *MockClient) DeleteAlarmNotificationProfile(ctx context.Context, n, s string) error {
	if m.DeleteAlarmNotificationProfileFunc != nil {
		return m.DeleteAlarmNotificationProfileFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetOpcUaConnection(ctx context.Context, n string) (*ResourceResponse[OpcUaConnectionConfig], error) {
	if m.GetOpcUaConnectionFunc != nil {
		return m.GetOpcUaConnectionFunc(ctx, n)
	}
	return &ResourceResponse[OpcUaConnectionConfig]{}, nil
}

func (m *MockClient) CreateOpcUaConnection(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error) {
	if m.CreateOpcUaConnectionFunc != nil {
		return m.CreateOpcUaConnectionFunc(ctx, i)
	}
	return &ResourceResponse[OpcUaConnectionConfig]{}, nil
}

func (m *MockClient) UpdateOpcUaConnection(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error) {
	if m.UpdateOpcUaConnectionFunc != nil {
		return m.UpdateOpcUaConnectionFunc(ctx, i)
	}
	return &ResourceResponse[OpcUaConnectionConfig]{}, nil
}

func (m *MockClient) DeleteOpcUaConnection(ctx context.Context, n, s string) error {
	if m.DeleteOpcUaConnectionFunc != nil {
		return m.DeleteOpcUaConnectionFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetAlarmJournal(ctx context.Context, n string) (*ResourceResponse[AlarmJournalConfig], error) {
	if m.GetAlarmJournalFunc != nil {
		return m.GetAlarmJournalFunc(ctx, n)
	}
	return &ResourceResponse[AlarmJournalConfig]{}, nil
}

func (m *MockClient) CreateAlarmJournal(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error) {
	if m.CreateAlarmJournalFunc != nil {
		return m.CreateAlarmJournalFunc(ctx, i)
	}
	return &ResourceResponse[AlarmJournalConfig]{}, nil
}

func (m *MockClient) UpdateAlarmJournal(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error) {
	if m.UpdateAlarmJournalFunc != nil {
		return m.UpdateAlarmJournalFunc(ctx, i)
	}
	return &ResourceResponse[AlarmJournalConfig]{}, nil
}

func (m *MockClient) DeleteAlarmJournal(ctx context.Context, n, s string) error {
	if m.DeleteAlarmJournalFunc != nil {
		return m.DeleteAlarmJournalFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetHistorianProvider(ctx context.Context, n string) (*ResourceResponse[HistorianProviderConfig], error) {
	if m.GetHistorianProviderFunc != nil {
		return m.GetHistorianProviderFunc(ctx, n)
	}
	return &ResourceResponse[HistorianProviderConfig]{}, nil
}

func (m *MockClient) CreateHistorianProvider(ctx context.Context, i ResourceResponse[HistorianProviderConfig]) (*ResourceResponse[HistorianProviderConfig], error) {
	if m.CreateHistorianProviderFunc != nil {
		return m.CreateHistorianProviderFunc(ctx, i)
	}
	return &ResourceResponse[HistorianProviderConfig]{}, nil
}

func (m *MockClient) UpdateHistorianProvider(ctx context.Context, i ResourceResponse[HistorianProviderConfig]) (*ResourceResponse[HistorianProviderConfig], error) {
	if m.UpdateHistorianProviderFunc != nil {
		return m.UpdateHistorianProviderFunc(ctx, i)
	}
	return &ResourceResponse[HistorianProviderConfig]{}, nil
}

func (m *MockClient) DeleteHistorianProvider(ctx context.Context, n, s string) error {
	if m.DeleteHistorianProviderFunc != nil {
		return m.DeleteHistorianProviderFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetSMTPProfile(ctx context.Context, n string) (*ResourceResponse[SMTPProfileConfig], error) {
	if m.GetSMTPProfileFunc != nil {
		return m.GetSMTPProfileFunc(ctx, n)
	}
	return &ResourceResponse[SMTPProfileConfig]{}, nil
}

func (m *MockClient) CreateSMTPProfile(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error) {
	if m.CreateSMTPProfileFunc != nil {
		return m.CreateSMTPProfileFunc(ctx, i)
	}
	return &ResourceResponse[SMTPProfileConfig]{}, nil
}

func (m *MockClient) UpdateSMTPProfile(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error) {
	if m.UpdateSMTPProfileFunc != nil {
		return m.UpdateSMTPProfileFunc(ctx, i)
	}
	return &ResourceResponse[SMTPProfileConfig]{}, nil
}

func (m *MockClient) DeleteSMTPProfile(ctx context.Context, n, s string) error {
	if m.DeleteSMTPProfileFunc != nil {
		return m.DeleteSMTPProfileFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetStoreAndForward(ctx context.Context, n string) (*ResourceResponse[StoreAndForwardConfig], error) {
	if m.GetStoreAndForwardFunc != nil {
		return m.GetStoreAndForwardFunc(ctx, n)
	}
	return &ResourceResponse[StoreAndForwardConfig]{}, nil
}

func (m *MockClient) CreateStoreAndForward(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error) {
	if m.CreateStoreAndForwardFunc != nil {
		return m.CreateStoreAndForwardFunc(ctx, i)
	}
	return &ResourceResponse[StoreAndForwardConfig]{}, nil
}

func (m *MockClient) UpdateStoreAndForward(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error) {
	if m.UpdateStoreAndForwardFunc != nil {
		return m.UpdateStoreAndForwardFunc(ctx, i)
	}
	return &ResourceResponse[StoreAndForwardConfig]{}, nil
}

func (m *MockClient) DeleteStoreAndForward(ctx context.Context, n, s string) error {
	if m.DeleteStoreAndForwardFunc != nil {
		return m.DeleteStoreAndForwardFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetIdentityProvider(ctx context.Context, n string) (*ResourceResponse[IdentityProviderConfig], error) {
	if m.GetIdentityProviderFunc != nil {
		return m.GetIdentityProviderFunc(ctx, n)
	}
	return &ResourceResponse[IdentityProviderConfig]{}, nil
}

func (m *MockClient) CreateIdentityProvider(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error) {
	if m.CreateIdentityProviderFunc != nil {
		return m.CreateIdentityProviderFunc(ctx, i)
	}
	return &ResourceResponse[IdentityProviderConfig]{}, nil
}

func (m *MockClient) UpdateIdentityProvider(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error) {
	if m.UpdateIdentityProviderFunc != nil {
		return m.UpdateIdentityProviderFunc(ctx, i)
	}
	return &ResourceResponse[IdentityProviderConfig]{}, nil
}

func (m *MockClient) DeleteIdentityProvider(ctx context.Context, n, s string) error {
	if m.DeleteIdentityProviderFunc != nil {
		return m.DeleteIdentityProviderFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetGanOutgoing(ctx context.Context, n string) (*ResourceResponse[GanOutgoingConfig], error) {
	if m.GetGanOutgoingFunc != nil {
		return m.GetGanOutgoingFunc(ctx, n)
	}
	return &ResourceResponse[GanOutgoingConfig]{}, nil
}

func (m *MockClient) CreateGanOutgoing(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error) {
	if m.CreateGanOutgoingFunc != nil {
		return m.CreateGanOutgoingFunc(ctx, i)
	}
	return &ResourceResponse[GanOutgoingConfig]{}, nil
}

func (m *MockClient) UpdateGanOutgoing(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error) {
	if m.UpdateGanOutgoingFunc != nil {
		return m.UpdateGanOutgoingFunc(ctx, i)
	}
	return &ResourceResponse[GanOutgoingConfig]{}, nil
}

func (m *MockClient) DeleteGanOutgoing(ctx context.Context, n, s string) error {
	if m.DeleteGanOutgoingFunc != nil {
		return m.DeleteGanOutgoingFunc(ctx, n, s)
	}
	return nil
}

func (m *MockClient) GetRedundancyConfig(ctx context.Context) (*RedundancyConfig, error) {
	if m.GetRedundancyConfigFunc != nil {
		return m.GetRedundancyConfigFunc(ctx)
	}
	return &RedundancyConfig{}, nil
}

func (m *MockClient) UpdateRedundancyConfig(ctx context.Context, c RedundancyConfig) error {
	if m.UpdateRedundancyConfigFunc != nil {
		return m.UpdateRedundancyConfigFunc(ctx, c)
	}
	return nil
}

func (m *MockClient) GetGanGeneralSettings(ctx context.Context) (*ResourceResponse[GanGeneralSettingsConfig], error) {
	if m.GetGanGeneralSettingsFunc != nil {
		return m.GetGanGeneralSettingsFunc(ctx)
	}
	return &ResourceResponse[GanGeneralSettingsConfig]{}, nil
}

func (m *MockClient) UpdateGanGeneralSettings(ctx context.Context, i ResourceResponse[GanGeneralSettingsConfig]) (*ResourceResponse[GanGeneralSettingsConfig], error) {
	if m.UpdateGanGeneralSettingsFunc != nil {
		return m.UpdateGanGeneralSettingsFunc(ctx, i)
	}
	return &ResourceResponse[GanGeneralSettingsConfig]{}, nil
}

func (m *MockClient) GetDevice(ctx context.Context, n string) (*ResourceResponse[DeviceConfig], error) {
	if m.GetDeviceFunc != nil {
		return m.GetDeviceFunc(ctx, n)
	}
	return &ResourceResponse[DeviceConfig]{}, nil
}

func (m *MockClient) CreateDevice(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error) {
	if m.CreateDeviceFunc != nil {
		return m.CreateDeviceFunc(ctx, i)
	}
	return &ResourceResponse[DeviceConfig]{}, nil
}

func (m *MockClient) UpdateDevice(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error) {
	if m.UpdateDeviceFunc != nil {
		return m.UpdateDeviceFunc(ctx, i)
	}
	return &ResourceResponse[DeviceConfig]{}, nil
}

func (m *MockClient) DeleteDevice(ctx context.Context, n, s string) error {
	if m.DeleteDeviceFunc != nil {
		return m.DeleteDeviceFunc(ctx, n, s)
	}
	return nil
}
