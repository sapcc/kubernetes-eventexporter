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

	eventRouter := NewEventRouter(clientset, eventsInformer, config)
	stop := sigHandler()

	go func() {
		glog.Info("Starting prometheus metrics")
		http.Handle("/metrics", promhttp.Handler())
		glog.Warning(http.ListenAndServe(metricsAddr, nil))
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

	if len(context) > 0 {
		overrides.CurrentContext = context
	}

	if len(kubeconfig) > 0 {
		rules.ExplicitPath = kubeconfig
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}
