package cruise_test

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/heptiolabs/cruise/http"
	"github.com/heptiolabs/cruise/internal/cruise"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeUptimeChecker struct {
	CreateUptimeCheckCalled  bool
	CreateUptimeCheckInError bool
	DeleteUptimeCheckCalled  bool
	DeleteUptimeCheckInError bool
	checks                   map[string]*http.UptimeCheck
}

func (f *fakeUptimeChecker) CreateUptimeCheck(check *http.UptimeCheck) error {
	f.CreateUptimeCheckCalled = true
	if f.CreateUptimeCheckInError {
		return fmt.Errorf("Something went wrong")
	}
	f.checks[check.Hostname] = check
	return nil
}

func (f *fakeUptimeChecker) DeleteUptimeCheck(hostname string) error {
	f.DeleteUptimeCheckCalled = true
	if f.DeleteUptimeCheckInError {
		return fmt.Errorf("Something went wrong")
	}
	delete(f.checks, hostname)
	return nil
}

func (f *fakeUptimeChecker) SyncUptimeChecks() error {
	return nil
}

func (f *fakeUptimeChecker) UptimeChecks() map[string]*http.UptimeCheck {
	return f.checks
}

func newFakeUptimeChecker() *fakeUptimeChecker {
	return &fakeUptimeChecker{
		checks: map[string]*http.UptimeCheck{},
	}
}

func newCruise(checker http.UptimeChecker) (*cruise.Cruise, *test.Hook) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	return cruise.NewCruise(checker, logger), hook
}

func TestOnAddNonIngress(t *testing.T) {
	f := newFakeUptimeChecker()
	c, log := newCruise(f)
	nonIngress := ""
	c.OnAdd(nonIngress)

	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Equal(t, "OnAdd unexpected type string: \"\"", log.LastEntry().Message)
}

func TestOnUpdateNonIngress(t *testing.T) {
	f := newFakeUptimeChecker()
	c, log := newCruise(f)
	nonIngress := ""
	c.OnUpdate(nonIngress, nonIngress)

	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Equal(t, "OnUpdate unexpected type string: \"\"", log.LastEntry().Message)
}

func TestOnUpdateOldNonIngress(t *testing.T) {
	f := newFakeUptimeChecker()
	c, log := newCruise(f)
	nonIngress := ""
	i := &v1beta1.Ingress{}
	c.OnUpdate(nonIngress, i)

	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Contains(t, log.LastEntry().Message, "received invalid oldObj *v1beta1.Ingress")
}

func TestOnDeleteNonIngress(t *testing.T) {
	f := newFakeUptimeChecker()
	c, log := newCruise(f)
	nonIngress := ""
	c.OnDelete(nonIngress)

	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Equal(t, "OnDelete unexpected type string: \"\"", log.LastEntry().Message)
}

func TestOnAddIngressWithNonExistingUptimeCheck(t *testing.T) {
	f := newFakeUptimeChecker()
	i := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mynamespace",
			Name:      "example",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)
	c.OnAdd(i)
	assert.True(t, f.CreateUptimeCheckCalled)

	check := &http.UptimeCheck{
		Hostname:               "example.com",
		Name:                   "mynamespace/example (example.com:80)",
		EnableTLS:              false,
		CheckIntervalInMinutes: 1,
	}

	assert.Equal(t, f.UptimeChecks()["example.com"], check)
}

func TestOnAddIngressWithErrorWhenCreatingUptimeCheck(t *testing.T) {
	f := newFakeUptimeChecker()
	f.CreateUptimeCheckInError = true

	i := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mynamespace",
			Name:      "example",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, log := newCruise(f)
	c.OnAdd(i)
	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Equal(t, "Something went wrong", log.LastEntry().Message)
	assert.Empty(t, f.UptimeChecks())
}

func TestOnDeleteIngressWithErrorWhenCreatingUptimeCheck(t *testing.T) {
	f := newFakeUptimeChecker()
	f.DeleteUptimeCheckInError = true

	i := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mynamespace",
			Name:      "example",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, log := newCruise(f)
	c.OnDelete(i)
	assert.Equal(t, logrus.ErrorLevel, log.LastEntry().Level)
	assert.Equal(t, "Something went wrong", log.LastEntry().Message)
	assert.Empty(t, f.UptimeChecks())
}

func TestOnAddIngressWithNonExistingUptimeCheckTLS(t *testing.T) {
	f := newFakeUptimeChecker()
	i := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mynamespace",
			Name:      "example",
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts: []string{"example.com"},
				},
			},
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)
	c.OnAdd(i)
	assert.True(t, f.CreateUptimeCheckCalled)

	check := &http.UptimeCheck{
		Hostname:               "example.com",
		Name:                   "mynamespace/example (example.com:443)",
		EnableTLS:              true,
		CheckIntervalInMinutes: 1,
	}

	assert.Equal(t, f.UptimeChecks()["example.com"], check)
}

func TestOnAddIngressWithoutHost(t *testing.T) {
	f := newFakeUptimeChecker()
	i := &v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{},
			},
		},
	}

	c, _ := newCruise(f)

	c.OnAdd(i)
	assert.False(t, f.CreateUptimeCheckCalled)
}

func TestOnAddIngressWithExistingUptimeCheck(t *testing.T) {
	f := &fakeUptimeChecker{
		checks: map[string]*http.UptimeCheck{
			"example.com": &http.UptimeCheck{},
		},
	}

	i := &v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)

	c.OnAdd(i)
	assert.False(t, f.CreateUptimeCheckCalled)
}

func TestOnDeleteIngress(t *testing.T) {
	f := &fakeUptimeChecker{
		checks: map[string]*http.UptimeCheck{
			"example.com": &http.UptimeCheck{},
		},
	}

	i := &v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)

	c.OnDelete(i)
	assert.True(t, f.DeleteUptimeCheckCalled)
	assert.Empty(t, f.UptimeChecks())
}

func TestOnUpdateIngressUptimeCheck(t *testing.T) {
	f := &fakeUptimeChecker{
		checks: map[string]*http.UptimeCheck{
			"example.com": &http.UptimeCheck{},
		},
	}

	old := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "mynamespace",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	new := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "mynamespace",
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts: []string{"example.com"},
				},
			},
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)

	c.OnUpdate(old, new)

	assert.True(t, f.DeleteUptimeCheckCalled)
	assert.True(t, f.CreateUptimeCheckCalled)

	check := &http.UptimeCheck{
		Hostname:               "example.com",
		Name:                   "mynamespace/example (example.com:443)",
		EnableTLS:              true,
		CheckIntervalInMinutes: 1,
	}

	assert.Equal(t, f.UptimeChecks()["example.com"], check)
}

func TestOnUpdateIngressWithNoOldHost(t *testing.T) {
	f := &fakeUptimeChecker{
		checks: map[string]*http.UptimeCheck{
			"example.com": &http.UptimeCheck{},
		},
	}

	old := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "mynamespace",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{},
			},
		},
	}

	new := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "mynamespace",
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts: []string{"example.com"},
				},
			},
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: "example.com",
				},
			},
		},
	}

	c, _ := newCruise(f)

	c.OnUpdate(old, new)

	assert.True(t, f.DeleteUptimeCheckCalled)
	assert.True(t, f.CreateUptimeCheckCalled)

	check := &http.UptimeCheck{
		Hostname:               "example.com",
		Name:                   "mynamespace/example (example.com:443)",
		EnableTLS:              true,
		CheckIntervalInMinutes: 1,
	}

	assert.Equal(t, f.UptimeChecks()["example.com"], check)
}
