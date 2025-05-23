package services

// Service is a type for the service part of a scope and/or role.
type Service string

// The list of registered services that publish scopes and/or roles.
const (
	CodyGateway      Service = "cody_gateway"
	Dotcom           Service = "dotcom"
	SAMS             Service = "sams"
	TelemetryGateway Service = "telemetry_gateway"
	EnterprisePortal Service = "enterprise_portal"
	MailGatekeeper   Service = "mail_gatekeeper"
	Workspaces       Service = "workspaces"
	SSC              Service = "ssc"
	Analytics        Service = "analytics"
)

var serviceNames = map[Service]string{
	CodyGateway:      "Cody Gateway",
	Dotcom:           "Sourcegraph Dotcom",
	SAMS:             "Sourcegraph Accounts Management System",
	TelemetryGateway: "Telemetry Gateway",
	EnterprisePortal: "Enterprise Portal",
	MailGatekeeper:   "Mail Gatekeeper",
	Workspaces:       "Workspaces",
	SSC:              "Self Serve Cody",
	Analytics:        "Sourcegraph Analytics",
}

func (s Service) DisplayName() string {
	if v, ok := serviceNames[s]; ok {
		return v
	}
	return string(s)
}
