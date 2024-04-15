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
	"errors"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"

	v1 "k8s.io/api/core/v1"
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
	Config         *Config
}

func NewEventRouter(kubeClient kubernetes.Interface, eventsInformer coreinformers.EventInformer, config *Config) (*EventRouter, error) {
	kubernetesEventCounterVec = make(map[string]*prometheus.CounterVec)

	for _, metric := range config.Metrics {
		var labels []string

		for key := range metric.Labels {
			labels = append(labels, key)
		}

		kubernetesEventCounterVec[metric.Name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metric.Name,
			Help: "Kubernetes Eventexporter Metric " + metric.Name,
		}, labels)

		prometheus.MustRegister(kubernetesEventCounterVec[metric.Name])
	}

	router := &EventRouter{
		kubeClient: kubeClient,
		Config:     config,
	}
	_, err := eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    router.addEvent,
		UpdateFunc: router.updateEvent,
		DeleteFunc: router.deleteEvent,
	})
	if err != nil {
		return nil, err
	}
	router.eLister = eventsInformer.Lister()
	router.eListerSynched = eventsInformer.Informer().HasSynced

	return router, err
}

func (er *EventRouter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer glog.Infof("Shutting down EventRouter")

	glog.Infof("Starting EventRouter")

	if !cache.WaitForCacheSync(stopCh, er.eListerSynched) {
		utilruntime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

func (er *EventRouter) addEvent(obj interface{}) {
	e, ok := obj.(*v1.Event)
	if !ok {
		glog.Warning("got non event from informer")
		return
	}

	if discardEvent(e) {
		glog.V(5).Infof("Discarding event: %v", e)
		return
	}

	filterMatches := LogEvent(e, er)
	for _, filterMatch := range filterMatches {
		prometheusEvent(filterMatch.Name, filterMatch.Labels)
	}
}

func (er *EventRouter) updateEvent(objOld, objNew interface{}) {
	eNew, ok := objNew.(*v1.Event)
	if !ok {
		glog.Warning("got non event from informer")
		return
	}

	if discardEvent(eNew) {
		glog.V(5).Infof("Discarding event: %v", eNew)
		return
	}

	filterMatches := LogEvent(eNew, er)
	for _, filterMatch := range filterMatches {
		prometheusEvent(filterMatch.Name, filterMatch.Labels)
	}
}

func prometheusEvent(filter string, labels map[string]string) {
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
	e, ok := obj.(*v1.Event)
	if !ok {
		glog.Warning("got non event from informer")
		return
	}
	glog.V(5).Infof("Event Deleted from the system:\n%v", e)
}

func discardEvent(e *v1.Event) bool {
	return time.Since(e.LastTimestamp.Time) > discardInterval
}
