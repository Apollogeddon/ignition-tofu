package client

import "fmt"

type APIErrorResponse struct {
	Success       bool     `json:"success"`
	Messages      []string `json:"messages,omitempty"`
	FieldMessages []struct {
		FieldName string   `json:"fieldName"`
		Messages  []string `json:"messages"`
	} `json:"fieldMessages,omitempty"`
	Problem *struct {
		Message    string   `json:"message"`
		StackTrace []string `json:"stacktrace,omitempty"`
	} `json:"problem,omitempty"`
}

func (e *APIErrorResponse) Error() string {
	var msg string
	switch {
	case e.Problem != nil:
		msg = "API error: " + e.Problem.Message
	case len(e.Messages) > 0:
		msg = fmt.Sprintf("API error: %v", e.Messages)
	default:
		msg = "unknown API error"
	}

	if len(e.FieldMessages) > 0 {
		msg += fmt.Sprintf(" (Field Errors: %v)", e.FieldMessages)
	}
	return msg
}

type ResourceResponse[T any] struct {
	Module  string `json:"module,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled,omitempty"`
	// sent even when empty, so a description removed from the configuration is cleared
	Description string `json:"description"`
	Signature   string `json:"signature,omitempty"`
	Config      T      `json:"config"`
}

type ResourceChangesResponse struct {
	Success bool `json:"success"`
	Changes []struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Collection   string `json:"collection"`
		NewSignature string `json:"newSignature"`
	} `json:"changes"`
}

type DatabaseConfig struct {
	Driver     string `json:"driver"`
	Translator string `json:"translator,omitempty"`
	ConnectURL string `json:"connectURL"`
	Username   string `json:"username,omitempty"`
	Password   any    `json:"password,omitempty"`
}

type TagProviderProfile struct {
	Type string `json:"type"`
}

type TagProviderConfig struct {
	Profile     TagProviderProfile `json:"profile"`
	Description string             `json:"description,omitempty"`
	Settings    map[string]any     `json:"settings"`
}

type UserSourceProfile struct {
	Type            string `json:"type"`
	FailoverProfile string `json:"failoverProfile,omitempty"`
	FailoverMode    string `json:"failoverMode,omitempty"`
	// nil leaves the gateway's value
	ScheduleRestricted *bool `json:"scheduleRestricted,omitempty"`
}

type UserSourceConfig struct {
	Profile UserSourceProfile `json:"profile"`
}

// Project's and the audit profile's optional booleans and numbers are pointers: nil leaves
// the gateway's value, while false or 0 is sent, which omitempty on a plain value would drop.
// A project's strings are always sent: an update keeps fields it leaves out, so an empty
// string is what clears one removed from the configuration.
type Project struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"description"`
	Title            string `json:"title"`
	Enabled          *bool  `json:"enabled,omitempty"`
	Parent           string `json:"parent"`
	Inheritable      *bool  `json:"inheritable,omitempty"`
	DefaultDB        string `json:"defaultDb"`
	TagProvider      string `json:"tagProvider"`
	UserSource       string `json:"userSource"`
	IdentityProvider string `json:"identityProvider"`
}

type AuditProfileSettings struct {
	DatabaseName          string `json:"databaseName,omitempty"`
	PruneEnabled          *bool  `json:"pruneEnabled,omitempty"`
	AutoCreate            *bool  `json:"autoCreate,omitempty"`
	TableName             string `json:"tableName,omitempty"`
	RemoteServer          string `json:"remoteServer,omitempty"`
	RemoteProfile         string `json:"remoteProfile,omitempty"`
	EnableStoreAndForward *bool  `json:"enableStoreAndForward,omitempty"`
}

type AuditProfileProfile struct {
	Type          string `json:"type"`
	RetentionDays *int   `json:"retentionDays,omitempty"`
}

type AuditProfileConfig struct {
	Profile  AuditProfileProfile  `json:"profile"`
	Settings AuditProfileSettings `json:"settings"`
}

type AlarmNotificationProfileProfile struct {
	Type string `json:"type"`
}

type AlarmNotificationProfileConfig struct {
	Profile  AlarmNotificationProfileProfile `json:"profile"`
	Settings map[string]any                  `json:"settings"`
}

type OpcUaConnectionEndpoint struct {
	DiscoveryURL   string `json:"discoveryUrl"`
	EndpointURL    string `json:"endpointUrl"`
	SecurityPolicy string `json:"securityPolicy"`
	SecurityMode   string `json:"securityMode"`
}

type OpcUaConnectionSettings struct {
	Endpoint OpcUaConnectionEndpoint `json:"endpoint"`
}

type OpcUaConnectionProfile struct {
	Type string `json:"type"`
}

type OpcUaConnectionConfig struct {
	Profile  OpcUaConnectionProfile  `json:"profile"`
	Settings OpcUaConnectionSettings `json:"settings"`
}

type AlarmJournalProfile struct {
	Type      string `json:"type"`
	QueryOnly bool   `json:"queryOnly,omitempty"`
}

type AlarmJournalSettings struct {
	Datasource    string `json:"datasource,omitempty"`
	RemoteGateway *struct {
		TargetServer  string `json:"targetServer,omitempty"`
		TargetJournal string `json:"targetJournal,omitempty"`
	} `json:"remoteGateway,omitempty"`
	Advanced *struct {
		TableName          string `json:"tableName,omitempty"`
		DataTableName      string `json:"dataTableName,omitempty"`
		UseStoreAndForward bool   `json:"useStoreAndForward,omitempty"`
	} `json:"advanced,omitempty"`
	Events *struct {
		MinPriority            string `json:"minPriority,omitempty"`
		StoreShelvedEvents     bool   `json:"storeShelvedEvents,omitempty"`
		StoreFromEnabledChange bool   `json:"storeFromEnabledChange,omitempty"`
	} `json:"events,omitempty"`
}

type AlarmJournalConfig struct {
	Profile  AlarmJournalProfile  `json:"profile"`
	Settings AlarmJournalSettings `json:"settings"`
}

type SMTPProfileSettingsClassic struct {
	Hostname        string `json:"hostname"`
	Port            int    `json:"port"`
	UseSslPort      bool   `json:"useSslPort"`
	StartTlsEnabled bool   `json:"startTlsEnabled"`
	Username        string `json:"username,omitempty"`
	Password        any    `json:"password,omitempty"`
}

type SMTPProfileSettings struct {
	Settings *SMTPProfileSettingsClassic `json:"settings,omitempty"`
}

type SMTPProfileProfile struct {
	Type string `json:"type"`
}

type SMTPProfileConfig struct {
	Profile  SMTPProfileProfile  `json:"profile"`
	Settings SMTPProfileSettings `json:"settings"`
}

type StoreAndForwardMaintenancePolicy struct {
	Action string `json:"action"`
	Limit  struct {
		LimitType string `json:"limitType"`
		Value     int    `json:"value"`
	} `json:"limit"`
}

type StoreAndForwardConfig struct {
	TimeThresholdMs            int                               `json:"timeThresholdMs"`
	ForwardRateMs              int                               `json:"forwardRateMs"`
	ForwardingPolicy           string                            `json:"forwardingPolicy"`
	ForwardingSchedule         string                            `json:"forwardingSchedule,omitempty"`
	IsThirdParty               bool                              `json:"isThirdParty"`
	DataThreshold              int                               `json:"dataThreshold"`
	BatchSize                  int                               `json:"batchSize"`
	ScanRateMs                 int                               `json:"scanRateMs"`
	PrimaryMaintenancePolicy   *StoreAndForwardMaintenancePolicy `json:"primaryMaintenancePolicy,omitempty"`
	SecondaryMaintenancePolicy *StoreAndForwardMaintenancePolicy `json:"secondaryMaintenancePolicy,omitempty"`
}

type IdentityProviderAuthMethod struct {
	Type   string `json:"type"`
	Config any    `json:"config"`
}

type IdentityProviderInternalConfig struct {
	UserSource               string                       `json:"userSource"`
	AuthMethods              []IdentityProviderAuthMethod `json:"authMethods"`
	SessionInactivityTimeout float64                      `json:"sessionInactivityTimeout"`
	SessionExp               float64                      `json:"sessionExp"`
	RememberMeExp            float64                      `json:"rememberMeExp"`
}

type IgnitionSecret struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type IdentityProviderOidcConfig struct {
	ClientId                   string          `json:"clientId"`
	ClientSecret               *IgnitionSecret `json:"clientSecret,omitempty"`
	ProviderId                 string          `json:"providerId"`
	AuthorizationEndpoint      string          `json:"authorizationEndpoint"`
	TokenEndpoint              string          `json:"tokenEndpoint"`
	JsonWebKeysEndpoint        string          `json:"jsonWebKeysEndpoint"`
	JsonWebKeysEndpointEnabled bool            `json:"jsonWebKeysEndpointEnabled"`
	UserInfoEndpoint           string          `json:"userInfoEndpoint,omitempty"`
	EndSessionEndpoint         string          `json:"endSessionEndpoint,omitempty"`
}

type IdentityProviderSamlConfig struct {
	IdpEntityId       string `json:"idpEntityId"`
	SpEntityId        string `json:"spEntityId,omitempty"`
	SpEntityIdEnabled bool   `json:"spEntityIdEnabled"`
	AcsBinding        string `json:"acsBinding"`
	NameIdFormat      string `json:"nameIdFormat"`
	SsoServiceConfig  struct {
		Uri     string `json:"uri"`
		Binding string `json:"binding"`
	} `json:"ssoServiceConfig"`
	ForceAuthnEnabled              bool     `json:"forceAuthnEnabled"`
	ResponseSignaturesRequired     bool     `json:"responseSignaturesRequired"`
	AssertionSignaturesRequired    bool     `json:"assertionSignaturesRequired"`
	IdpMetadataUrl                 string   `json:"idpMetadataUrl,omitempty"`
	IdpMetadataUrlEnabled          bool     `json:"idpMetadataUrlEnabled"`
	SignatureVerifyingCertificates []string `json:"signatureVerifyingCertificates"`
	SignatureVerifyingKeys         []any    `json:"signatureVerifyingKeys"`
}

type IdentityProviderSecurityLevelRules struct {
	Nodes []any `json:"nodes"`
}

type IdentityProviderProfile struct {
	Type                string                             `json:"type"`
	UserAttributeMapper map[string]any                     `json:"userAttributeMapper"`
	UserGrants          map[string]any                     `json:"userGrants"`
	SecurityLevelRules  IdentityProviderSecurityLevelRules `json:"securityLevelRules"`
}

type IdentityProviderConfig struct {
	Profile  IdentityProviderProfile `json:"profile"`
	Settings any                     `json:"settings"`
}

type GanOutgoingConfig struct {
	Host                     string  `json:"host"`
	Port                     int     `json:"port"`
	UseSSL                   bool    `json:"useSSL"`
	PingRateMillis           float64 `json:"pingRateMillis,omitempty"`
	PingTimeoutMillis        float64 `json:"pingTimeoutMillis,omitempty"`
	PingMaxMissed            float64 `json:"pingMaxMissed,omitempty"`
	WsTimeoutMillis          float64 `json:"wsTimeoutMillis,omitempty"`
	HttpConnectTimeoutMillis float64 `json:"httpConnectTimeoutMillis,omitempty"`
	HttpReadTimeoutMillis    float64 `json:"httpReadTimeoutMillis,omitempty"`
	SendThreads              float64 `json:"sendThreads,omitempty"`
	ReceiveThreads           float64 `json:"receiveThreads,omitempty"`
}

type RedundancyConfig struct {
	Role                string `json:"role"`
	ActiveHistoryLevel  string `json:"activeHistoryLevel"`
	JoinWaitTime        int    `json:"joinWaitTime"`
	RecoveryMode        string `json:"masterRecoveryMode"`
	AllowHistoryCleanup bool   `json:"allowHistoryCleanup"`
	GatewayNetworkSetup *struct {
		Host               string  `json:"host,omitempty"`
		Port               int     `json:"port,omitempty"`
		EnableSsl          bool    `json:"enableSsl"`
		PingRate           float64 `json:"pingRate,omitempty"`
		PingTimeout        float64 `json:"pingTimeout,omitempty"`
		PingMaxMissed      float64 `json:"pingMaxMissed,omitempty"`
		WebsocketTimeout   float64 `json:"websocketTimeout,omitempty"`
		HttpConnectTimeout float64 `json:"httpConnectTimeout,omitempty"`
		HttpReadTimeout    float64 `json:"httpReadTimeout,omitempty"`
		SendThreads        float64 `json:"sendThreads,omitempty"`
		ReceiveThreads     float64 `json:"receiveThreads,omitempty"`
	} `json:"gatewayNetworkSetup,omitempty"`
}

type GanGeneralSettingsConfig struct {
	RequireSSL                  bool    `json:"requireSSL"`
	RequireTwoWayAuth           bool    `json:"requireTwoWayAuth"`
	AllowIncoming               bool    `json:"allowIncoming"`
	SecurityPolicy              string  `json:"securityPolicy"`
	Whitelist                   string  `json:"whitelist,omitempty"`
	AllowedProxyHops            float64 `json:"allowedProxyHops"`
	WebsocketSessionIdleTimeout float64 `json:"websocketSessionIdleTimeout"`
	TempFilesMaxAgeHours        float64 `json:"tempFilesMaxAgeHours"`
}

type DeviceConfig map[string]any

// DeviceWireProfile and DeviceWireConfig describe the actual on-the-wire
// shape the gateway expects for device resources: driver-specific settings
// nested under a profile/settings envelope, keyed by driver type. DeviceConfig
// itself stays a flat map since that's what the rest of the provider (and its
// tests) work with; the client layer wraps/unwraps at the HTTP boundary.
type DeviceWireProfile struct {
	Type string `json:"type"`
}

type DeviceWireConfig struct {
	Profile  DeviceWireProfile `json:"profile"`
	Settings DeviceConfig      `json:"settings"`
}
