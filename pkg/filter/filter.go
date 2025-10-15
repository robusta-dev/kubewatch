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
	"reflect"
	"strconv"
	"strings"

	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/sirupsen/logrus"

	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	events_v1 "k8s.io/api/events/v1"
)

// Filter is the main filter struct
type Filter struct {
	enabled bool
}

// NewFilter creates a new filter instance
func NewFilter() *Filter {
	enabled := false
	if envVal := os.Getenv("ADVANCED_FILTERS"); envVal != "" {
		parsedVal, err := strconv.ParseBool(envVal)
		if err == nil {
			enabled = parsedVal
		} else {
			logrus.Warnf("Invalid ADVANCED_FILTERS value: %s, defaulting to false", envVal)
		}
	}

	if enabled {
		logrus.Info("Advanced filtering is ENABLED")
	} else {
		logrus.Info("Advanced filtering is DISABLED")
	}

	return &Filter{
		enabled: enabled,
	}
}

// ShouldSendEvent determines if an event should be sent to Robusta
func (f *Filter) ShouldSendEvent(e event.Event) bool {
	// If filtering is disabled, send all events
	if !f.enabled {
		return true
	}

	// Apply filtering rules based on resource kind
	switch e.Kind {
	case "Event":
		return f.shouldSendEventResource(e)
	case "Job":
		return f.shouldSendJobEvent(e)
	case "Pod":
		return f.shouldSendPodEvent(e)
	default:
		// For all other resources, send the event
		return true
	}
}

// shouldSendEventResource filters Kubernetes Event resources
func (f *Filter) shouldSendEventResource(e event.Event) bool {
	// For Event resources, only send warning events and only create events
	if e.Reason != "Created" {
		logrus.Debugf("Filtering out Event resource - reason: %s (only 'Created' events are sent)", e.Reason)
		return false
	}

	// Check the event reason - always send Evicted events regardless of type
	isEvictedEvent := false
	switch obj := e.Obj.(type) {
	case *api_v1.Event:
		if obj.Reason == "Evicted" {
			isEvictedEvent = true
			logrus.Debugf("Event resource with reason 'Evicted' will be sent regardless of type")
		}
	case *events_v1.Event:
		if obj.Reason == "Evicted" {
			isEvictedEvent = true
			logrus.Debugf("Event resource with reason 'Evicted' will be sent regardless of type")
		}
	}

	if isEvictedEvent {
		return true
	}

	// Check if it's a warning event
	switch obj := e.Obj.(type) {
	case *api_v1.Event:
		if obj.Type != api_v1.EventTypeWarning {
			logrus.Debugf("Filtering out Event resource - type: %s (only Warning events are sent)", obj.Type)
			return false
		}
	case *events_v1.Event:
		if obj.Type != api_v1.EventTypeWarning {
			logrus.Debugf("Filtering out Event resource - type: %s (only Warning events are sent)", obj.Type)
			return false
		}
	default:
		// If we can't determine the type, send it to be safe
		logrus.Warnf("Unable to determine Event type for filtering, sending event")
		return true
	}

	return true
}

// shouldSendJobEvent filters Job events
func (f *Filter) shouldSendJobEvent(e event.Event) bool {
	// Always send Create and Delete events
	if e.Reason == "Created" || e.Reason == "Deleted" {
		return true
	}

	// For Update events, check if spec changed or job failed
	if e.Reason == "Updated" {
		job, ok := e.Obj.(*batch_v1.Job)
		if !ok {
			logrus.Warnf("Unable to cast Job object for filtering, sending event")
			return true
		}

		oldJob, ok := e.OldObj.(*batch_v1.Job)
		if !ok {
			// If we don't have the old object, send the event to be safe
			return true
		}

		// Check if spec changed
		if !reflect.DeepEqual(job.Spec, oldJob.Spec) {
			logrus.Debugf("Job %s spec changed, sending update event", job.Name)
			return true
		}

		// Check if job failed
		for _, condition := range job.Status.Conditions {
			if condition.Type == batch_v1.JobFailed && condition.Status == api_v1.ConditionTrue {
				logrus.Debugf("Job %s failed, sending update event", job.Name)
				return true
			}
		}

		logrus.Debugf("Filtering out Job update event - no spec change or failure detected")
		return false
	}

	// For other event types, don't send
	return false
}

// shouldSendPodEvent filters Pod events
func (f *Filter) shouldSendPodEvent(e event.Event) bool {
	// Always send Create and Delete events
	if e.Reason == "Created" || e.Reason == "Deleted" {
		return true
	}

	// For Update events, apply specific filters
	if e.Reason == "Updated" {
		pod, ok := e.Obj.(*api_v1.Pod)
		if !ok {
			logrus.Warnf("Unable to cast Pod object for filtering, sending event")
			return true
		}

		oldPod, ok := e.OldObj.(*api_v1.Pod)
		if !ok {
			// If we don't have the old object, send the event to be safe
			return true
		}

		// Check if spec changed
		if !reflect.DeepEqual(pod.Spec, oldPod.Spec) {
			logrus.Debugf("Pod %s spec changed, sending update event", pod.Name)
			return true
		}

		// Check for container restarts
		if f.hasContainerRestarted(pod) {
			logrus.Debugf("Pod %s has container restarts, sending update event", pod.Name)
			return true
		}

		// Check for ImagePullBackOff
		if f.hasImagePullBackOff(pod) {
			logrus.Debugf("Pod %s has ImagePullBackOff, sending update event", pod.Name)
			return true
		}

		// Check if pod is evicted
		if f.isPodEvicted(pod) {
			logrus.Debugf("Pod %s is evicted, sending update event", pod.Name)
			return true
		}

		// Check for OOMKilled
		if f.hasOOMKilled(pod) {
			logrus.Debugf("Pod %s has OOMKilled container, sending update event", pod.Name)
			return true
		}

		logrus.Debugf("Filtering out Pod update event - no significant changes detected")
		return false
	}

	// For other event types, don't send
	return false
}

// hasContainerRestarted checks if any container (including init containers) has restarted
func (f *Filter) hasContainerRestarted(pod *api_v1.Pod) bool {
	// Check init containers
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.RestartCount > 0 {
			return true
		}
	}

	// Check regular containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.RestartCount > 0 {
			return true
		}
	}

	return false
}

// hasImagePullBackOff checks if any container is waiting due to ImagePullBackOff
func (f *Filter) hasImagePullBackOff(pod *api_v1.Pod) bool {
	// Check init containers
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.State.Waiting != nil &&
			containerStatus.State.Waiting.Reason == "ImagePullBackOff" {
			return true
		}
	}

	// Check regular containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil &&
			containerStatus.State.Waiting.Reason == "ImagePullBackOff" {
			return true
		}
	}

	return false
}

// isPodEvicted checks if the pod has been evicted
func (f *Filter) isPodEvicted(pod *api_v1.Pod) bool {
	// Check pod phase and reason
	if pod.Status.Phase == api_v1.PodFailed && pod.Status.Reason == "Evicted" {
		return true
	}

	// Alternative check: look for eviction in pod status message
	if strings.Contains(pod.Status.Message, "evicted") {
		return true
	}

	return false
}

// hasOOMKilled checks if any container was OOMKilled
func (f *Filter) hasOOMKilled(pod *api_v1.Pod) bool {
	// Check init containers
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if f.isContainerOOMKilled(containerStatus) {
			return true
		}
	}

	// Check regular containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if f.isContainerOOMKilled(containerStatus) {
			return true
		}
	}

	return false
}

// isContainerOOMKilled checks if a specific container status indicates OOMKilled
func (f *Filter) isContainerOOMKilled(status api_v1.ContainerStatus) bool {
	// Check current state
	if status.State.Terminated != nil &&
		status.State.Terminated.Reason == "OOMKilled" {
		return true
	}

	// Check last state
	if status.LastTerminationState.Terminated != nil &&
		status.LastTerminationState.Terminated.Reason == "OOMKilled" {
		return true
	}

	return false
}
