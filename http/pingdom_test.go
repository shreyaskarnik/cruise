package http_test

import (
	"os"
	"testing"

	"github.com/heptiolabs/cruise/http"
	"github.com/stretchr/testify/assert"
)

var (
	pingdomLiveTest bool
	username        string
	password        string
	apikey          string
)

func init() {
	username = os.Getenv("PINGDOM_USERNAME")
	password = os.Getenv("PINGDOM_PASSWORD")
	apikey = os.Getenv("PINGDOM_APIKEY")

	if len(username) > 0 && len(password) > 0 && len(apikey) > 0 {
		pingdomLiveTest = true
	}
}

func TestPingdomUptimeChecker(t *testing.T) {
	if !pingdomLiveTest {
		t.Skip("skipping live test")
	}

	c, _ := http.NewPindomUptimeChecker(username, password, apikey)

	check := &http.UptimeCheck{
		Hostname:               "google.com",
		Name:                   "mynamespace / google (google.com:443)",
		EnableTLS:              true,
		CheckIntervalInMinutes: 1,
	}

	err := c.CreateUptimeCheck(check)
	assert.Nil(t, err)
	assert.Equal(t, "google.com", c.UptimeChecks()["google.com"].Hostname)
	assert.Equal(t, "mynamespace / google (google.com:443)", c.UptimeChecks()["google.com"].Name)
	assert.Equal(t, 1, c.UptimeChecks()["google.com"].CheckIntervalInMinutes)
	assert.True(t, c.UptimeChecks()["google.com"].EnableTLS)
	assert.NotEqual(t, "", c.UptimeChecks()["google.com"].ID)

	n, _ := http.NewPindomUptimeChecker(username, password, apikey)
	err = n.SyncUptimeChecks()
	assert.Nil(t, err)
	assert.Equal(t, "google.com", c.UptimeChecks()["google.com"].Hostname)
	assert.Equal(t, "mynamespace / google (google.com:443)", c.UptimeChecks()["google.com"].Name)
	assert.Equal(t, 1, c.UptimeChecks()["google.com"].CheckIntervalInMinutes)
	assert.True(t, c.UptimeChecks()["google.com"].EnableTLS)
	assert.NotEqual(t, "", c.UptimeChecks()["google.com"].ID)

	err = n.DeleteUptimeCheck(check.Hostname)
	assert.Nil(t, err)
	assert.Nil(t, n.UptimeChecks()["google.com"])
}
