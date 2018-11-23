package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	yaml "gopkg.in/yaml.v2"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Metrics []struct {
		Name         string `yaml:"name"`
		EventMatcher []struct {
			Key  string `yaml:"key"`
			Expr string `yaml:"expr"`
		} `yaml:"event_matcher"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metrics"`
}

var (
	configFile     string
	metricsAddr    string
	kubeconfigFile string
	kubeContext    string
)

func init() {
	flag.StringVar(&configFile, "config", "/etc/eventexporter/config.yaml", "config file for the event exporter")
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
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point - this is why we can't have nice things
		sig := <-c
		glog.Warningf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

func loadConfig() (Config, error) {
	yamlFile, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	defer yamlFile.Close()

	byteValue, _ := ioutil.ReadAll(yamlFile)

	var config Config
	err = yaml.Unmarshal([]byte(byteValue), &config)

	if err != nil {
		glog.Fatalf("Could not unmarshal config: %v", err)
	}

	return config, nil
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	config, err := loadConfig()

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
