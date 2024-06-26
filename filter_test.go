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
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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
)

func TestLogEventMatch(t *testing.T) {
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	testEvent := v1.Event{
		Message:        "Created container",
		InvolvedObject: v1.ObjectReference{Kind: "Pod"},
		Source: v1.EventSource{
			Host: "Testnode",
		},
		Type: "Normal",
	}
	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Equal(t, []FilterMatch{
		{
			Name:   "metric_1",
			Labels: map[string]string{"node": testEvent.Source.Host, "type": testEvent.Type},
		},
		{
			Name:   "metric_2",
			Labels: map[string]string{"type": testEvent.Type},
		},
	}, matches)
}

func TestNoMatch(t *testing.T) {
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	testEvent := v1.Event{
		Message: "Other message",
		Source: v1.EventSource{
			Host: "Testnode",
		},
		Type: "Normal",
	}

	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Empty(t, matches)
}

func TestSkipMetricsWithMissingLabels(t *testing.T) {
	testConfig := []byte(`metrics:
- name: metric_1
  event_matcher:
  - key: Message
    expr: Label missing
  labels:
    missing: Nase
    type: Type
- name: metric_2
  event_matcher:
  - key: Message
    expr: Label missing
  labels:
    type: Type
`)
	testEvent := v1.Event{
		Message: "Label missing",
		Type:    "Normal",
	}
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Equal(t, []FilterMatch{
		{
			Name:   "metric_2",
			Labels: map[string]string{"type": testEvent.Type},
		},
	}, matches)
}

func TestLabelSubmatch(t *testing.T) {
	testConfig := []byte(`metrics:
- name: submatch
  event_matcher:
  - key: Message
    expr: Volume (.*) mount failed for Instance (.*)
  - key: Type
  expr: Normal
  labels:
    volume: Message[1]
    instance: Message[2]
`)
	testEvent := v1.Event{
		Message: "Volume vol-1234 mount failed for Instance instance-789",
		Type:    "Normal",
	}
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	matches := LogEvent(&testEvent, &EventRouter{Config: config})

	require.Equal(t, []FilterMatch{
		{
			Name:   "submatch",
			Labels: map[string]string{"volume": "vol-1234", "instance": "instance-789"},
		},
	}, matches)
}

func TestObjectReference(t *testing.T) {
	testConfig := []byte(`metrics:
- name: submatch
  event_matcher:
  - key: Type
    expr: Normal
  labels:
    node: Object.Spec.NodeName
`)
	config, err := NewConfig(bytes.NewBuffer(testConfig))
	require.NoError(t, err, "There should be no error while unmarshaling config")

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Spec: v1.PodSpec{NodeName: "test-node"},
	}

	fakeClient := fake.NewSimpleClientset(pod)
	event := v1.Event{
		InvolvedObject: v1.ObjectReference{
			Kind:       "Pod",
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			APIVersion: "v1",
		},
		Type: "Normal",
	}

	matches := LogEvent(&event, &EventRouter{Config: config, kubeClient: fakeClient})
	require.Equal(t, []FilterMatch{
		{Name: "submatch", Labels: map[string]string{"node": pod.Spec.NodeName}},
	}, matches)
}

func TestConfigErrorSubmatchWithoutMatcher(t *testing.T) {
	testConfig := []byte(`metrics:
- name: submatch
  event_matcher:
  - key: Type
  expr: Normal
  labels:
    volume: Message[1]
`)
	_, err := NewConfig(bytes.NewBuffer(testConfig))
	require.EqualError(t, err, "configuration for metric 'submatch' invalid: Can't use a submatch for key 'Message' without a match expression")
}
func TestConfigErrorSubmatchGroupMissing(t *testing.T) {
	testConfig := []byte(`metrics:
- name: submatch
  event_matcher:
  - key: Message
    expr: Normal
  labels:
    volume: Message[1]
`)
	_, err := NewConfig(bytes.NewBuffer(testConfig))
	require.EqualError(t, err, "configuration for metric 'submatch' invalid: Match expression for key 'Message' does not contain 1 subexpressions")
}

func TestLogEventEmptyConfig(t *testing.T) {
	matches := LogEvent(&v1.Event{}, &EventRouter{})

	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}

func TestLogEventEmptyEvent(t *testing.T) {
	testEvent := v1.Event{
		Message:        "Created container",
		InvolvedObject: v1.ObjectReference{Kind: "Pod"},
		Source: v1.EventSource{
			Host: "Testnode",
		},
		Type: "Normal",
	}
	matches := LogEvent(&testEvent, &EventRouter{})
	require.Equal(t, 0, len(matches), "There should be no metrics returned")
}
