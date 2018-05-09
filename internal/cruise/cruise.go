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
	"reflect"

	"github.com/heptiolabs/cruise/internal/pingdom"
	"github.com/sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
)

type Cruise struct {
	logger  logrus.FieldLogger
	checker pingdom.UptimeChecker
}

func NewCruise(checker pingdom.UptimeChecker, logger logrus.FieldLogger) *Cruise {
	return &Cruise{
		logger:  logger,
		checker: checker,
	}
}

func (c *Cruise) OnAdd(obj interface{}) {
	ing, ok := obj.(*v1beta1.Ingress)
	if !ok {
		c.logger.Errorf("OnAdd unexpected type %T: %#v", obj, obj)
	}
	c.recompute(nil, ing)
}

func (c *Cruise) OnUpdate(oldObj, newObj interface{}) {
	switch newObj := newObj.(type) {
	case *v1beta1.Ingress:
		oldObj, ok := oldObj.(*v1beta1.Ingress)
		if !ok {
			c.logger.Errorf("OnUpdate endpoints %#v received invalid oldObj %T; %#v", newObj, oldObj, oldObj)
			return
		}
		c.recompute(oldObj, newObj)
	default:
		c.logger.Errorf("OnUpdate unexpected type %T: %#v", newObj, newObj)
	}
}

func (c *Cruise) OnDelete(obj interface{}) {
	switch obj := obj.(type) {
	case *v1beta1.Ingress:
		c.recompute(obj, nil)
	default:
		c.logger.Errorf("OnDelete unexpected type %T: %#v", obj, obj)
	}
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

	// store a list of active hostnames, anything present in olding but missing from
	// newing will be removed.
	active := make(map[string]bool)

	for _, r := range newing.Spec.Rules {
		host := r.Host
		if host == "" {
			c.logger.WithField("ingress", fmt.Sprintf("%s/%s", newing.Namespace, newing.Name)).Debugf("skipping rule %d, missing Host field", r.IngressRuleValue)
			continue
		}

		active[host] = true // mark this host as active even if we end up skipping it

		if _, ok := c.checker.UptimeChecks()[host]; ok {
			if olding.ObjectMeta.Name == "" ||
				(reflect.DeepEqual(olding.Spec.Rules, newing.Spec.Rules) && reflect.DeepEqual(olding.Spec.TLS, newing.Spec.Rules)) {
				c.logger.WithField("hostname", host).Info("check already exists, skipping")
				continue
			} else {
				c.checker.DeleteUptimeCheck(host)
			}
		}

		port := 80
		if newing.Spec.TLS != nil {
			port = 443
		}

		check := pingdom.UptimeCheck{
			Name:                   fmt.Sprintf("%s/%s (%s:%d)", newing.Namespace, newing.Name, host, port),
			Hostname:               host,
			CheckIntervalInMinutes: 1,
			EnableTLS:              port == 443,
		}

		err := c.checker.CreateUptimeCheck(&check)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		c.logger.Info("check created")
	}

	for _, r := range olding.Spec.Rules {
		host := r.Host
		if host == "" {
			c.logger.Debugf("skipping rule %d, missing Host field", r.IngressRuleValue)
			continue
		}

		if active[host] {
			// do not remove, this is an active check
			continue
		}

		err := c.checker.DeleteUptimeCheck(host)
		if err != nil {
			c.logger.WithField("hostname", host).Error(err)
			continue
		}

		c.logger.Info("check deleted")
	}
}
