package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	kubernetesEventCounterVec map[string]*prometheus.CounterVec
)

type EventRouter struct {
	kubeClient     kubernetes.Interface
	eLister        corelisters.EventLister
	eListerSynched cache.InformerSynced
	Config         Config
}

func NewEventRouter(kubeClient kubernetes.Interface, eventsInformer coreinformers.EventInformer, config Config) *EventRouter {
	kubernetesEventCounterVec = make(map[string]*prometheus.CounterVec)

	for _, metric := range config.Metrics {
		var labels []string

		for key := range metric.Labels {
			labels = append(labels, key)
		}

		kubernetesEventCounterVec[metric.Name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metric.Name,
			Help: fmt.Sprintf("Kubernetes Eventexporter Metric %s", metric.Name),
		}, labels)

		prometheus.MustRegister(kubernetesEventCounterVec[metric.Name])
	}

	er := &EventRouter{
		kubeClient: kubeClient,
		Config:     config,
	}
	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    er.addEvent,
		UpdateFunc: er.updateEvent,
		DeleteFunc: er.deleteEvent,
	})
	er.eLister = eventsInformer.Lister()
	er.eListerSynched = eventsInformer.Informer().HasSynced

	return er
}

func (er *EventRouter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer glog.Infof("Shutting down EventRouter")

	glog.Infof("Starting EventRouter")

	if !cache.WaitForCacheSync(stopCh, er.eListerSynched) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

func (er *EventRouter) addEvent(obj interface{}) {
	e := obj.(*v1.Event)

	if discardEvent(e) {
		glog.V(5).Infof("Discarding event: %v", e)
		return
	}

	filterMatches := LogEvent(e, er)
	if filterMatches != nil {
		for _, filterMatch := range filterMatches {
			prometheusEvent(e, filterMatch.Name, filterMatch.Label)
		}
	}
}

func (er *EventRouter) updateEvent(objOld interface{}, objNew interface{}) {
	eNew := objNew.(*v1.Event)

	if discardEvent(eNew) {
		glog.V(5).Infof("Discarding event: %v", eNew)
		return
	}

	filterMatches := LogEvent(eNew, er)
	if filterMatches != nil {
		for _, filterMatch := range filterMatches {
			prometheusEvent(eNew, filterMatch.Name, filterMatch.Label)
		}
	}
}

func prometheusEvent(event *v1.Event, filter string, labels map[string]string) {
	var counter prometheus.Counter
	var err error

	glog.V(5).Infof("Sending labels: %v", labels)

	counter, err = kubernetesEventCounterVec[filter].GetMetricWith(labels)

	if err != nil {
		glog.Warning(err)
	} else {
		counter.Add(1)
	}
}

func (er *EventRouter) deleteEvent(obj interface{}) {
	e := obj.(*v1.Event)
	glog.V(5).Infof("Event Deleted from the system:\n%v", e)
}

func discardEvent(e *v1.Event) bool {
	return time.Since(e.LastTimestamp.Time) > discardInterval
}
