package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

func TestLogEventMatch(t *testing.T) {
	var config Config
	var event *v1.Event

	configYAML, eventJSON := getTestData()

	err := yaml.Unmarshal([]byte(configYAML), &config)
	require.NoError(t, err, "There should be no error while unmarshaling config")

	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err, "There should be no error while unmarshaling event")

	matches := LogEvent(event, &EventRouter{Config: config})

	require.Equal(t, 2, len(matches), "There should be exactly two metrics returned")
	require.Equal(t, "metric_1", matches[0].Name)
	require.Equal(t, "Testnode", matches[0].Label["node"])
	require.Equal(t, "Normal", matches[0].Label["type"])
	require.Equal(t, "metric_2", matches[1].Name)
	require.Equal(t, "Normal", matches[1].Label["type"])
}

func TestLogEventEmptyConfig(t *testing.T) {
	var config Config

	configYAML, _ := getTestData()

	err := yaml.Unmarshal([]byte(configYAML), &config)
	require.NoError(t, err, "There should be no error while unmarshaling config")

	matches := LogEvent(&v1.Event{}, &EventRouter{})

	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}

func TestLogEventEmptyEvent(t *testing.T) {
	var event *v1.Event

	_, eventJSON := getTestData()

	err := json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err, "There should be no error while unmarshaling event")

	matches := LogEvent(event, &EventRouter{})

	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}

func getTestData() (string, string) {
	config := `metrics:
- name: metric_1
  event_matcher:
  - key: InvolvedObject.Kind
    expr: Pod
  - key: Message
    expr: .*Created container.*
  labels:
    node: Source.Host
    type: Type
- name: metric_2
  event_matcher:
  - key: Message
    expr: .*Created container.*
  labels:
    type: Type`

	event := `{
		"Message": "Created container",
		"InvolvedObject": {
			"Kind": "Pod"
		},
		"Source": {
			"Host": "Testnode"	
		},
		"Type": "Normal"
	}`

	return config, event
}
