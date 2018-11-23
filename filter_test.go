package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
)

func TestLogEvent(t *testing.T) {
	var config Config
	var event *v1.Event

	configJSON, eventJSON := getTestData()

	err := json.Unmarshal([]byte(configJSON), &config)
	require.NoError(t, err, "There should be no error while unmarshaling config")

	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err, "There should be no error while unmarshaling event")

	matches := LogEvent(event, config)

	require.Equal(t, "metric_1", matches[0].Name)
	require.Equal(t, []string{"Testnode"}, matches[0].Label)
	require.Equal(t, "metric_2", matches[1].Name)
	require.Equal(t, []string{"Normal"}, matches[1].Label)
}

func getTestData() (string, string) {
	config := `{
		"metrics": [
		  {
			"name": "metric_1",
			"event_filters": [
			  {
				"key": "InvolvedObject.Kind",
				"expr": "Pod"
			  },
			  {
				"key": "Message",
				"expr": ".*Created container.*"
			  }
			],
			"labels": [
			  {
				"label": "Source.Host"
			  }
			]
		  },
		  {
			"name": "metric_2",
			"event_filters": [
			  {
				"key": "Source.Host",
				"expr": "Testnode"
			  }
			],
			"labels": [
			  {
				"label": "Reason"
			  }
			]
		  }
		]
	  }`

	event := `{
		"Message": "Created container",
		"InvolvedObject": {
			"Kind": "Pod"
		},
		"Source": {
			"Host": "Testnode"	
		},
		"Reason": "Normal"
	}`

	return config, event
}
