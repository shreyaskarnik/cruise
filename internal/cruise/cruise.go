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

// Package cruise contains the business logic that listens ingrewss objects
// and translates those into calls to the pingdom api.
package cruise

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type Cruise struct {
	logrus.FieldLogger
}

func (c *Cruise) OnAdd(obj interface{}) {
	switch obj := obj.(type) {
	case *v1beta1.Ingress:
		_ = obj
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
		_, _ = oldObj, newObj
	default:
		c.Errorf("OnUpdate unexpected type %T: %#v", newObj, newObj)
	}
}

func (c *Cruise) OnDelete(obj interface{}) {
	switch obj := obj.(type) {
	case *v1beta1.Ingress:
		_ = obj
	case cache.DeletedFinalStateUnknown:
		c.OnDelete(obj.Obj) // recurse into ourselves with the tombstoned value
	default:
		c.Errorf("OnDelete unexpected type %T: %#v", obj, obj)
	}
}
