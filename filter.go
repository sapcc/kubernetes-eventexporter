package main

import (
	"fmt"
	"regexp"
	"strings"

	structs_util "github.com/fatih/structs"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type FilterMatch struct {
	Name  string
	Label map[string]string
}

func LogEvent(event *v1.Event, config Config) []FilterMatch {
	var matches []FilterMatch

	for _, metric := range config.Metrics {
		ret := true

		for _, filter := range metric.EventMatcher {
			value, err := GetValueFromStruct(event, filter.Key)

			if filter.Expr != "" && err == nil {
				match, _ := regexp.MatchString(
					filter.Expr,
					value,
				)

				glog.V(5).Infof("Expression: %s Value: %s Match: %v\n", filter.Expr, value, match)

				ret = ret && match
			}

			if err != nil {
				ret = false
				break
			}
		}

		if ret {
			var l = make(map[string]string)

			for labelKey, labelSpec := range metric.Labels {
				value, err := GetValueFromStruct(event, labelSpec)

				if err != nil {
					break
				}

				l[labelKey] = value
			}

			matches = append(matches, FilterMatch{Name: metric.Name, Label: l})
		}
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
