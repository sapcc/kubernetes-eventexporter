package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

var (
	testConfig = []byte(`metrics:
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
    type: Type
`)
	testEvent = v1.Event{
		Message: "Created container",
		InvolvedObject: v1.ObjectReference{
			Kind: "Pod",
		},
		Source: v1.EventSource{
			Host: "Testnode",
		},
		Type: "Normal",
	}
)

func TestLogEventMatch(t *testing.T) {
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Equal(t, matches, []FilterMatch{
		FilterMatch{
			Name:   "metric_1",
			Labels: map[string]string{"node": testEvent.Source.Host, "type": testEvent.Type},
		},
		FilterMatch{
			Name:   "metric_2",
			Labels: map[string]string{"type": testEvent.Type},
		},
	})
}

func TestNoMatch(t *testing.T) {
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	testEvent = v1.Event{
		Message: "Other message",
		Source: v1.EventSource{
			Host: "Testnode",
		},
		Type: "Normal",
	}

	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Empty(t, matches)
}

func TestLogEventEmptyConfig(t *testing.T) {
	matches := LogEvent(&v1.Event{}, &EventRouter{})

	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}

func TestLogEventEmptyEvent(t *testing.T) {

	matches := LogEvent(&testEvent, &EventRouter{})

	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}
