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
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

var (
	labelSubMatchRE = regexp.MustCompile(`^(.*)\[([0-9])\]$`)
)

type LookupFunc = func(event *v1.Event, matches map[string][]string) (string, error)

type Config struct {
	Metrics []struct {
		Name         string `yaml:"name"`
		EventMatcher []struct {
			Key  string `yaml:"key"`
			Expr string `yaml:"expr"`
		} `yaml:"event_matcher"`
		Labels         map[string]string `yaml:"labels"`
		regexMap       map[string]*regexp.Regexp
		labelLookupMap map[string]LookupFunc
	} `yaml:"metrics"`
}

func NewConfig(reader io.Reader) (*Config, error) {
	var config Config
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	for i, metric := range config.Metrics {
		config.Metrics[i].regexMap = make(map[string]*regexp.Regexp, len(metric.EventMatcher))
		for _, matcher := range metric.EventMatcher {
			r, err := regexp.Compile(matcher.Expr)
			if err != nil {
				return nil, fmt.Errorf("configuration for metric '%s' invalid: match expression for key %s invalid: %w", metric.Name, matcher.Key, err)
			}
			if _, found := config.Metrics[i].regexMap[matcher.Key]; found {
				return nil, fmt.Errorf("configuration for metric '%s' invalid: Multiple matchers for key '%s'", metric.Name, matcher.Key)
			}
			config.Metrics[i].regexMap[matcher.Key] = r
		}
		config.Metrics[i].labelLookupMap = make(map[string]func(*v1.Event, map[string][]string) (string, error), len(metric.Labels))

		// create lookup map for label values
		for key, l := range metric.Labels {
			labelSpec := l // local copy of l, needed for closures
			if strings.HasPrefix(labelSpec, PodVirtualTypePrefix) {
				config.Metrics[i].labelLookupMap[key] = func(event *v1.Event, _ map[string][]string) (string, error) {
					pod, err := getPodObjectForEvent(event)
					if err != nil {
						return "", err
					}
					return GetValueFromStruct(pod, strings.TrimPrefix(labelSpec, PodVirtualTypePrefix))
				}
			} else {
				if matches := labelSubMatchRE.FindStringSubmatch(labelSpec); matches != nil {
					label := matches[1]
					submatch, err := strconv.Atoi(matches[2])
					if err != nil {
						return nil, fmt.Errorf("failed to parse label %s for metric %s: %w", labelSpec, metric.Name, err)
					}
					re, found := config.Metrics[i].regexMap[label]
					if !found {
						return nil, fmt.Errorf("configuration for metric '%s' invalid: Can't use a submatch for key '%s' without a match expression", metric.Name, label)
					}
					if re.NumSubexp() < submatch {
						return nil, fmt.Errorf("configuration for metric '%s' invalid: Match expression for key '%s' does not contain %d subexpressions", metric.Name, label, submatch)
					}
					config.Metrics[i].labelLookupMap[key] = func(_ *v1.Event, matches map[string][]string) (string, error) { //nolint:unparam
						return matches[label][submatch], nil
					}
				} else {
					config.Metrics[i].labelLookupMap[key] = func(event *v1.Event, _ map[string][]string) (string, error) {
						return GetValueFromStruct(event, labelSpec)
					}
				}
			}
		}
	}

	return &config, nil
}
