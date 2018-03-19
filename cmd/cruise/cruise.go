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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/heptio/cruise/internal/cruise"
	"github.com/heptio/cruise/internal/k8s"
	"github.com/heptio/cruise/internal/workgroup"

	"github.com/sirupsen/logrus"
)

// this is necessary due to #113 wherein glog neccessitates a call to flag.Parse
// before any logging statements can be invoked. (See also https://github.com/golang/glog/blob/master/glog.go#L679)
// unsure why this seemingly unnecessary prerequisite is in place but there must be some sane reason.
func init() {
	flag.Parse()
}

func main() {
	log := logrus.StandardLogger()
	c := &cruise.Cruise{
		FieldLogger: log,
	}

	app := kingpin.New("cruise", "Remote HTTP monitoring operator.")

	serve := app.Command("serve", "Serve xDS API traffic")
	inCluster := serve.Flag("incluster", "use in cluster configuration.").Bool()
	kubeconfig := serve.Flag("kubeconfig", "path to kubeconfig (if not in running inside a cluster)").Default(filepath.Join(os.Getenv("HOME"), ".kube", "config")).String()

	args := os.Args[1:]
	switch kingpin.MustParse(app.Parse(args)) {
	default:
		app.Usage(args)
		os.Exit(2)
	case serve.FullCommand():
		log.Infof("args: %v", args)
		var g workgroup.Group

		// buffer notifications to t to ensure they are handled sequentially.
		buf := k8s.NewBuffer(&g, c, log, 128)

		client := newClient(*kubeconfig, *inCluster)

		wl := log.WithField("context", "watch")
		k8s.WatchIngress(&g, client, wl, buf)

		g.Run()
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

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
