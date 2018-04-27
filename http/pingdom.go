package http

import (
	"fmt"
	"regexp"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

type PingdomUptimeChecker struct {
	userID       int
	client       *pingdom.Client
	uptimeChecks map[string]*UptimeCheck
}

func NewPindomUptimeChecker(user, password, key string) (UptimeChecker, error) {
	client := pingdom.NewClient(user, password, key)

	// refresh contact list and locate the userid of c.Client.User
	// because of a limitation in the 2.0 api we have to pick the first
	// contact id and hope it's the billing contact.
	contacts, err := client.Contacts.List()
	if err != nil {
		return nil, err
	}

	if len(contacts) < 1 {
		return nil, fmt.Errorf("cannot locate user id for Client.User %q", client.User)
	}

	c := &PingdomUptimeChecker{
		userID:       contacts[0].ID,
		client:       client,
		uptimeChecks: make(map[string]*UptimeCheck),
	}

	err = c.SyncUptimeChecks()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *PingdomUptimeChecker) UptimeChecks() map[string]*UptimeCheck {
	return c.uptimeChecks
}

func (c *PingdomUptimeChecker) SyncUptimeChecks() error {
	list, err := c.client.Checks.List()
	if err != nil {
		return err
	}
	for _, pc := range list {
		c.uptimeChecks[pc.Hostname] = toUptimeCheck(pc)
	}
	return nil
}

func (c *PingdomUptimeChecker) CreateUptimeCheck(check *UptimeCheck) error {
	pc := pingdom.HttpCheck{
		Name:                     check.Name,
		Hostname:                 check.Hostname,
		Resolution:               check.CheckIntervalInMinutes,
		Encryption:               check.EnableTLS,
		SendNotificationWhenDown: 1, // TODO(dfc) no idea what this does, but the API barks if it is not set.
		ContactIds:               []int{c.userID},
	}

	res, err := c.client.Checks.Create(&pc)
	if err == nil {
		check.ID = res.ID
		c.uptimeChecks[check.Hostname] = check
	}

	return err
}

func (c *PingdomUptimeChecker) DeleteUptimeCheck(hostName string) error {
	check, exists := c.uptimeChecks[hostName]
	if !exists {
		return nil
	}

	_, err := c.client.Checks.Delete(check.ID)
	if err != nil {
		return err
	}

	delete(c.uptimeChecks, hostName)

	return nil
}

func toUptimeCheck(c pingdom.CheckResponse) *UptimeCheck {
	rp := regexp.MustCompile("443")
	return &UptimeCheck{
		Hostname: c.Hostname,
		ID:       c.ID,
		Name:     c.Name,
		CheckIntervalInMinutes: c.Resolution,
		EnableTLS:              rp.MatchString(c.Name), // Pingdom API does not show it so we need to rely on the name
	}
}
