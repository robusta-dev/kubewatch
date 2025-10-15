/*
Copyright 2024

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"os"
	"testing"

	"github.com/bitnami-labs/kubewatch/pkg/event"
	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	events_v1 "k8s.io/api/events/v1"
)

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"Disabled by default", "", false},
		{"Enabled with true", "true", true},
		{"Disabled with false", "false", false},
		{"Invalid value defaults to false", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ADVANCED_FILTERS", tt.envValue)
				defer os.Unsetenv("ADVANCED_FILTERS")
			}

			filter := NewFilter()
			if filter.enabled != tt.expected {
				t.Errorf("Expected enabled=%v, got %v", tt.expected, filter.enabled)
			}
		})
	}
}

func TestShouldSendEventResource(t *testing.T) {
	filter := &Filter{enabled: true}

	tests := []struct {
		name     string
		event    event.Event
		expected bool
	}{
		{
			name: "Warning Event Created - Should Send",
			event: event.Event{
				Kind:   "Event",
				Reason: "Created",
				Obj: &api_v1.Event{
					Type: api_v1.EventTypeWarning,
				},
			},
			expected: true,
		},
		{
			name: "Normal Event Created - Should Filter",
			event: event.Event{
				Kind:   "Event",
				Reason: "Created",
				Obj: &api_v1.Event{
					Type: api_v1.EventTypeNormal,
				},
			},
			expected: false,
		},
		{
			name: "Warning Event Updated - Should Filter",
			event: event.Event{
				Kind:   "Event",
				Reason: "Updated",
				Obj: &api_v1.Event{
					Type: api_v1.EventTypeWarning,
				},
			},
			expected: false,
		},
		{
			name: "Warning Event Deleted - Should Filter",
			event: event.Event{
				Kind:   "Event",
				Reason: "Deleted",
				Obj: &api_v1.Event{
					Type: api_v1.EventTypeWarning,
				},
			},
			expected: false,
		},
		{
			name: "Warning EventsV1 Created - Should Send",
			event: event.Event{
				Kind:   "Event",
				Reason: "Created",
				Obj: &events_v1.Event{
					Type: api_v1.EventTypeWarning,
				},
			},
			expected: true,
		},
		{
			name: "Evicted Event Normal Type - Should Send",
			event: event.Event{
				Kind:   "Event",
				Reason: "Created",
				Obj: &api_v1.Event{
					Type:   api_v1.EventTypeNormal,
					Reason: "Evicted",
				},
			},
			expected: true,
		},
		{
			name: "Evicted EventsV1 Normal Type - Should Send",
			event: event.Event{
				Kind:   "Event",
				Reason: "Created",
				Obj: &events_v1.Event{
					Type:   api_v1.EventTypeNormal,
					Reason: "Evicted",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldSendEvent(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldSendJobEvent(t *testing.T) {
	filter := &Filter{enabled: true}

	jobSpec1 := batch_v1.JobSpec{
		Parallelism: intPtr(1),
	}
	jobSpec2 := batch_v1.JobSpec{
		Parallelism: intPtr(2),
	}

	tests := []struct {
		name     string
		event    event.Event
		expected bool
	}{
		{
			name: "Job Created - Should Send",
			event: event.Event{
				Kind:   "Job",
				Reason: "Created",
				Obj:    &batch_v1.Job{},
			},
			expected: true,
		},
		{
			name: "Job Deleted - Should Send",
			event: event.Event{
				Kind:   "Job",
				Reason: "Deleted",
				Obj:    &batch_v1.Job{},
			},
			expected: true,
		},
		{
			name: "Job Updated with Spec Change - Should Send",
			event: event.Event{
				Kind:   "Job",
				Reason: "Updated",
				Obj: &batch_v1.Job{
					Spec: jobSpec2,
				},
				OldObj: &batch_v1.Job{
					Spec: jobSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Job Updated with Failure - Should Send",
			event: event.Event{
				Kind:   "Job",
				Reason: "Updated",
				Obj: &batch_v1.Job{
					Spec: jobSpec1,
					Status: batch_v1.JobStatus{
						Conditions: []batch_v1.JobCondition{
							{
								Type:   batch_v1.JobFailed,
								Status: api_v1.ConditionTrue,
							},
						},
					},
				},
				OldObj: &batch_v1.Job{
					Spec: jobSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Job Updated without Changes - Should Filter",
			event: event.Event{
				Kind:   "Job",
				Reason: "Updated",
				Obj: &batch_v1.Job{
					Spec: jobSpec1,
				},
				OldObj: &batch_v1.Job{
					Spec: jobSpec1,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldSendEvent(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldSendPodEvent(t *testing.T) {
	filter := &Filter{enabled: true}

	podSpec1 := api_v1.PodSpec{
		RestartPolicy: api_v1.RestartPolicyAlways,
	}
	podSpec2 := api_v1.PodSpec{
		RestartPolicy: api_v1.RestartPolicyNever,
	}

	tests := []struct {
		name     string
		event    event.Event
		expected bool
	}{
		{
			name: "Pod Created - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Created",
				Obj:    &api_v1.Pod{},
			},
			expected: true,
		},
		{
			name: "Pod Deleted - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Deleted",
				Obj:    &api_v1.Pod{},
			},
			expected: true,
		},
		{
			name: "Pod Updated with Spec Change - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec2,
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Pod Updated with Container Restart - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec1,
					Status: api_v1.PodStatus{
						ContainerStatuses: []api_v1.ContainerStatus{
							{
								RestartCount: 1,
							},
						},
					},
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Pod Updated with ImagePullBackOff - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec1,
					Status: api_v1.PodStatus{
						ContainerStatuses: []api_v1.ContainerStatus{
							{
								State: api_v1.ContainerState{
									Waiting: &api_v1.ContainerStateWaiting{
										Reason: "ImagePullBackOff",
									},
								},
							},
						},
					},
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Pod Evicted - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec1,
					Status: api_v1.PodStatus{
						Phase:  api_v1.PodFailed,
						Reason: "Evicted",
					},
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Pod with OOMKilled Container - Should Send",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec1,
					Status: api_v1.PodStatus{
						ContainerStatuses: []api_v1.ContainerStatus{
							{
								State: api_v1.ContainerState{
									Terminated: &api_v1.ContainerStateTerminated{
										Reason: "OOMKilled",
									},
								},
							},
						},
					},
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: true,
		},
		{
			name: "Pod Updated without Significant Changes - Should Filter",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
				Obj: &api_v1.Pod{
					Spec: podSpec1,
					Status: api_v1.PodStatus{
						Phase: api_v1.PodRunning,
					},
				},
				OldObj: &api_v1.Pod{
					Spec: podSpec1,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldSendEvent(tt.event)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldSendEventWithFilterDisabled(t *testing.T) {
	filter := &Filter{enabled: false}

	tests := []struct {
		name  string
		event event.Event
	}{
		{
			name: "Any Event",
			event: event.Event{
				Kind:   "Event",
				Reason: "Updated",
			},
		},
		{
			name: "Any Pod",
			event: event.Event{
				Kind:   "Pod",
				Reason: "Updated",
			},
		},
		{
			name: "Any Job",
			event: event.Event{
				Kind:   "Job",
				Reason: "Updated",
			},
		},
		{
			name: "Any Other Resource",
			event: event.Event{
				Kind:   "Deployment",
				Reason: "Updated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldSendEvent(tt.event)
			if !result {
				t.Errorf("Expected all events to be sent when filter is disabled")
			}
		})
	}
}

// Helper function to create int pointers
func intPtr(i int32) *int32 {
	return &i
}
