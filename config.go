package main

import (
	"fmt"
	"io"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Metrics []struct {
		Name         string `yaml:"name"`
		EventMatcher []struct {
			Key   string `yaml:"key"`
			Expr  string `yaml:"expr"`
			regex *regexp.Regexp
		} `yaml:"event_matcher"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metrics"`
}

func NewConfig(reader io.Reader) (*Config, error) {

	var config Config
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return nil, fmt.Errorf("Failed to parse config: %v", err)
	}
	for i, metric := range config.Metrics {
		for n, matcher := range metric.EventMatcher {
			r, err := regexp.Compile(matcher.Expr)
			if err != nil {
				return nil, fmt.Errorf("Regex for metric %s, key %s invalid: %s", metric.Name, matcher.Key, err)
			}
			config.Metrics[i].EventMatcher[n].regex = r
		}
	}

	return &config, nil
}
