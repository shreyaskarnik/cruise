package http

type UptimeCheck struct {
	Hostname               string
	Name                   string
	EnableTLS              bool
	CheckIntervalInMinutes int
	ID                     int
}

type UptimeChecker interface {
	UptimeChecks() map[string]*UptimeCheck
	SyncUptimeChecks() error
	CreateUptimeCheck(check *UptimeCheck) error
	DeleteUptimeCheck(hostName string) error
}
