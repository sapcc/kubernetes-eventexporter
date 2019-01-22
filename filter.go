package main

import (
	"fmt"
	"strings"

	structs_util "github.com/fatih/structs"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PodVirtualTypePrefix = "Object."
)

var (
	eventRouter *EventRouter
)

type FilterMatch struct {
	Name   string
	Labels map[string]string
}

func LogEvent(event *v1.Event, er *EventRouter) []FilterMatch {
	var matches []FilterMatch
	eventRouter = er
	if er.Config == nil {
		return matches
	}

OUTER:
	for _, metric := range er.Config.Metrics {

		for _, filter := range metric.EventMatcher {
			value, err := GetValueFromStruct(event, filter.Key)
			if err != nil {
				glog.Errorf("Could not get value for key %s: %v", filter.Key, err)
				continue OUTER
			}

			if filter.Expr != "" && err == nil {
				match := filter.regex.MatchString(value)
				glog.V(5).Infof("Expression: %s Value: %s Match: %v\n", filter.Expr, value, match)
				if !match {
					continue OUTER
				}
			}
		}

		var l = make(map[string]string)

		for labelKey, labelSpec := range metric.Labels {
			value := ""
			var err error
			var pod *v1.Pod

			if strings.HasPrefix(labelSpec, PodVirtualTypePrefix) && event.InvolvedObject.Kind == "Pod" {
				pod, err = getPodObjectForEvent(event)
				if err == nil {
					value, err = GetValueFromStruct(pod, strings.Replace(labelSpec, PodVirtualTypePrefix, "", 1))
				}
			} else {
				value, err = GetValueFromStruct(event, labelSpec)
			}

			if err != nil {
				glog.Errorf("Could not get label '%s' for metric '%s': %v", labelKey, metric.Name, err)
				continue OUTER
			}

			l[labelKey] = value
		}

		matches = append(matches, FilterMatch{Name: metric.Name, Labels: l})
	}

	return matches
}

func GetValueFromStruct(object interface{}, key string) (string, error) {
	keySlice := strings.Split(key, ".")
	s := structs_util.New(object)
	var newS *structs_util.Field
	ok := false

	for i, v := range keySlice {
		if i == 0 {
			newS, ok = s.FieldOk(v)
		} else {
			newS, ok = newS.FieldOk(v)
		}

		if !ok {
			return "", fmt.Errorf("Extracting value failed at %s, index %d", v, i)
		}
	}

	ret, ok := newS.Value().(string)

	if !ok {
		return "", fmt.Errorf("Value is not a string")
	}

	return ret, nil
}

func getPodObjectForEvent(event *v1.Event) (*v1.Pod, error) {
	return eventRouter.kubeClient.CoreV1().Pods(event.InvolvedObject.Namespace).Get(event.InvolvedObject.Name, metav1.GetOptions{})
}
