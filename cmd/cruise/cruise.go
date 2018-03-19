// Copyright © 2017 Heptio
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

// Cruise automates the creation of HTTP status checks for ingress resources.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/heptiolabs/cruise/internal/cruise"
	"github.com/russellcardullo/go-pingdom/pingdom"

	"github.com/sirupsen/logrus"
)

func init() {
	// thanks, glog
	flag.Parse()
}

func main() {
	log := logrus.StandardLogger()
	app := kingpin.New("cruise", "Remote HTTP monitoring operator.")

	serve := app.Command("serve", "Serve xDS API traffic")
	inCluster := serve.Flag("incluster", "use in cluster configuration.").Bool()
	kubeconfig := serve.Flag("kubeconfig", "path to kubeconfig (if not in running inside a cluster)").Default(filepath.Join(os.Getenv("HOME"), ".kube", "config")).String()
	username := serve.Flag("username", "Pingdom Username").Default(os.Getenv("PINGDOM_USERNAME")).String()
	password := serve.Flag("password", "Pingdom Password").Default(os.Getenv("PINGDOM_PASSWORD")).String()
	apikey := serve.Flag("apikey", "Pingdom API Key").Default(os.Getenv("PINGDOM_APIKEY")).String()

	args := os.Args[1:]
	switch kingpin.MustParse(app.Parse(args)) {
	default:
		app.Usage(args)
		os.Exit(2)
	case serve.FullCommand():
		log.Infof("args: %v", args)

		client := newClient(*kubeconfig, *inCluster)
		c := &cruise.Cruise{
			FieldLogger: log.WithField("context", "cruise"),
			Client:      pingdom.NewClient(*username, *password, *apikey),
		}
		w := watchIngress(client, c)
		w.Run(nil)
	}
}

func newClient(kubeconfig string, inCluster bool) *kubernetes.Clientset {
	var err error
	var config *rest.Config
	if kubeconfig != "" && !inCluster {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		check(err)
	} else {
		config, err = rest.InClusterConfig()
		check(err)
	}

	client, err := kubernetes.NewForConfig(config)
	check(err)
	return client
}

func watchIngress(client *kubernetes.Clientset, rs ...cache.ResourceEventHandler) cache.SharedInformer {
	lw := cache.NewListWatchFromClient(client.ExtensionsV1beta1().RESTClient(), "ingresses", v1.NamespaceAll, fields.Everything())
	sw := cache.NewSharedInformer(lw, new(v1beta1.Ingress), 30*time.Minute)
	for _, r := range rs {
		sw.AddEventHandler(r)
	}
	return sw
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
