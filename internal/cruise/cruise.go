// Copyright © 2018 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cruise contains the business logic that listens for ingress
// objects and translates those into calls to the pingdom api.
package cruise

import (
	"fmt"
	"sync"

	"github.com/russellcardullo/go-pingdom/pingdom"
	"github.com/sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type Cruise struct {
	logrus.FieldLogger

	mu sync.Mutex
	*pingdom.Client
	userId int // userid matching the PINGDOM_USERNAME field
	checks map[string]*pingdom.CheckResponse
}

func (c *Cruise) OnAdd(obj interface{}) {
	switch obj := obj.(type) {
	case *v1beta1.Ingress:
		c.recompute(nil, obj)
	default:
		c.Errorf("OnAdd unexpected type %T: %#v", obj, obj)
	}
}

func (c *Cruise) OnUpdate(oldObj, newObj interface{}) {
	switch newObj := newObj.(type) {
	case *v1beta1.Ingress:
		oldObj, ok := oldObj.(*v1beta1.Ingress)
		if !ok {
			c.Errorf("OnUpdate endpoints %#v received invalid oldObj %T; %#v", newObj, oldObj, oldObj)
			return
		}
		c.recompute(oldObj, newObj)
	default:
		c.Errorf("OnUpdate unexpected type %T: %#v", newObj, newObj)
	}
}

func (c *Cruise) OnDelete(obj interface{}) {
	switch obj := obj.(type) {
	case *v1beta1.Ingress:
		c.recompute(obj, nil)
	case cache.DeletedFinalStateUnknown:
		c.OnDelete(obj.Obj) // recurse into ourselves with the tombstoned value
	default:
		c.Errorf("OnDelete unexpected type %T: %#v", obj, obj)
	}
}

// client returns an active pingdom Client.
// On the first call to client the list of existing checks in c.checks is
// populated.
func (c *Cruise) client() *pingdom.Client {
	c.mu.Lock()

	// 1. refresh contact list and locate the userid of c.Client.User
	// because of a limitation in the 2.0 api we have to pick the first
	// contact id and hope it's the billing contact.
	contacts, err := c.Client.Contacts.List()
	if err != nil {
		c.Fatalf("cannot list existing contacts: %v", err)
		return nil // not reached
	}
	if len(contacts) < 1 {
		c.Fatalf("cannot locate user id for Client.User %q", c.Client.User)
		return nil // not reached
	}
	c.userId = contacts[0].ID

	// 2. populate check list
	if c.checks == nil {
		list, err := c.Client.Checks.List()
		if err != nil {
			c.Fatalf("cannot list existing checks: %v", err)
			return nil // not reached
		}
		c.checks = make(map[string]*pingdom.CheckResponse)
		for i := range list {
			check := list[i]
			c.checks[check.Hostname] = &check
		}
	}
	c.mu.Unlock()
	return c.Client
}

// recompute creates checks for hosts present in newing but missing from olding,
// and removes checks for hosts present in olding, but missing from newing.
func (c *Cruise) recompute(olding, newing *v1beta1.Ingress) {
	if olding == newing {
		// if olding/newing == nil or are the same object, skip
		return
	}

	// normalise old and new ingress objects; a nil object becomes a blank object of the same name
	if olding == nil {
		olding = &v1beta1.Ingress{
			ObjectMeta: newing.ObjectMeta,
		}
	}

	if newing == nil {
		newing = &v1beta1.Ingress{
			ObjectMeta: olding.ObjectMeta,
		}
	}

	log := c.WithField("ingress", fmt.Sprintf("%s/%s", newing.Namespace, newing.Name))

	_ = c.client() // grab client to make sure c.pc and c.checks are populated

	// store a list of active hostnames, anything present in olding but missing from
	// newing will be removed.
	active := make(map[string]bool)

	for i, r := range newing.Spec.Rules {
		host := r.Host
		if host == "" {
			log.Debugf("skipping rule %d, missing Host field", i)
			continue
		}
		active[host] = true // mark this host as active even if we end up skipping it

		log := log.WithField("hostname", host)
		if _, ok := c.checks[host]; ok {
			log.Info("check already exists, skipping")
			continue
		}

		port := 80
		if newing.Spec.TLS != nil {
			port = 443
		}
		check := pingdom.HttpCheck{
			Name:                     fmt.Sprintf("%s/%s (%s:%d)", newing.Namespace, newing.Name, host, port),
			Hostname:                 host,
			Resolution:               1, // check every minute
			Encryption:               port == 443,
			SendNotificationWhenDown: 1, // TODO(dfc) no idea what this does, but the API barks if it is not set.
			ContactIds:               []int{c.userId},
		}
		res, err := c.client().Checks.Create(&check)
		if err != nil {
			log.Error(err)
			continue
		}
		c.checks[check.Hostname] = res
		log.Infof("check created")
	}

	for i, r := range olding.Spec.Rules {
		host := r.Host
		if host == "" {
			log.Debugf("skipping rule %d, missing Host field", i)
			continue
		}
		if active[host] {
			// do not remove, this is an active check
			continue
		}
		log := log.WithField("hostname", host)
		check, ok := c.checks[host]
		if !ok {
			log.Errorf("cannot remove check, no cached entry") // can't remove the check without its ID
			continue
		}
		c.client().Checks.Delete(check.ID)
		delete(c.checks, host)
		log.Info("check deleted")
	}

}
