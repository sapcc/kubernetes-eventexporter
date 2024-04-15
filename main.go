// Copyright 2024 SAP SE
//
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
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	configFile      string
	metricsAddr     string
	kubeconfigFile  string
	kubeContext     string
	discardInterval time.Duration
)

func init() {
	flag.StringVar(&configFile, "config", "/etc/eventexporter/config.yaml", "config file for the event exporter")
	flag.DurationVar(&discardInterval, "discard", 60*time.Second, "Discard events older then specified Interaval. Set to 0 to disable")
	flag.StringVar(&metricsAddr, "listen-address", ":9102", "The address to listen on for HTTP requests.")
	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "Use explicit kubeconfig file")
	flag.StringVar(&kubeContext, "context", "", "Use context")
}

func sigHandler() <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM) // Termination Request
		sig := <-c
		glog.Warningf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	yamlFile, err := os.Open(configFile)
	if err != nil {
		glog.Fatalf("Failed to open file %s: %v", configFile, err)
	}
	config, err := NewConfig(yamlFile)
	yamlFile.Close()
	if err != nil {
		glog.Fatal("Could not load config file", err)
	}

	kubeconfig, err := kubeConfig(kubeconfigFile, kubeContext)

	if err != nil {
		glog.Fatal("Failed to create kubeconfig", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		glog.Fatal("Could not create client set", err)
	}

	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute*30)
	eventsInformer := sharedInformers.Core().V1().Events()

	eventRouter, err := NewEventRouter(clientset, eventsInformer, config)
	if err != nil {
		glog.Fatal("Failed to create event router: %s", err)
	}
	stop := sigHandler()

	go func() {
		glog.Info("Starting prometheus metrics")
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              metricsAddr,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           mux,
		}
		glog.Warning(server.ListenAndServe())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		eventRouter.Run(stop)
	}()

	glog.Infof("Starting shared Informer(s)")
	sharedInformers.Start(stop)
	wg.Wait()
	glog.Warningf("Exiting main()")
	os.Exit(1)
}

func kubeConfig(kubeconfig, context string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	if context != "" {
		overrides.CurrentContext = context
	}

	if kubeconfig != "" {
		rules.ExplicitPath = kubeconfig
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}
