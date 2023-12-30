/*
Copyright 2018 Bitnami

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

package cloudevent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

var cloudEventErrMsg = `
%s

You need to set Cloudevents webhook url
using "--url/-u" or using environment variables:

export KW_CLOUDEVENT_URL=webhook_url

Command line flags will override environment variables

`

// Webhook handler implements handler.Handler interface,
// Notify event to Webhook channel
type CloudEvent struct {
	Url       string
	StartTime uint64
	Counter   uint64

	cloudeventsClient cloudevents.Client
}

// EventMeta containes the meta data about the event occurred
type CloudEventMessageData struct {
	Operation   string         `json:"operation"`
	Kind        string         `json:"kind"`
	ClusterUid  string         `json:"clusterUid"`
	Description string         `json:"description"`
	ApiVersion  string         `json:"apiVersion"`
	Obj         runtime.Object `json:"obj"`
	OldObj      runtime.Object `json:"oldObj"`
}

func (m *CloudEvent) Init(c *config.Config) error {
	m.Url = c.Handler.CloudEvent.Url
	m.StartTime = uint64(time.Now().Unix())
	m.Counter = 0

	if m.Url == "" {
		m.Url = os.Getenv("KW_CLOUDEVENT_URL")
	}

	if m.Url == "" {
		return fmt.Errorf(cloudEventErrMsg, "Missing cloudevent url")
	}

	var err error
	m.cloudeventsClient, err = cloudevents.NewClientHTTP()
	if err != nil {
		return fmt.Errorf("failed to create client, %v", err)
	}

	return nil
}

func (m *CloudEvent) Handle(e event.Event) {
	m.Counter++ // TODO: do we have to worry about threadsafety here?

	event := cloudevents.NewEvent()
	event.SetSource("https://github.com/robusta-dev/kubewatch")
	event.SetType("KUBERNETES_TOPOLOGY_CHANGE")
	event.SetTime(time.Now())
	event.SetID(fmt.Sprintf("%v-%v", m.StartTime, m.Counter))
	if dataAssignmentError := event.SetData(cloudevents.ApplicationJSON, m.prepareMessage(e)); dataAssignmentError != nil {
		logrus.Printf("Failed to set data: %v", dataAssignmentError)
		return
	}

	result := m.cloudeventsClient.Send(cloudevents.ContextWithTarget(context.Background(), m.Url), event)
	if cloudevents.IsNACK(result) || cloudevents.IsUndelivered(result) {
		logrus.Printf("Failed to send: %v", result)
		return
	}

	logrus.Printf("Message successfully sent to %s at %s ", m.Url, time.Now())
}

func (m *CloudEvent) prepareMessage(e event.Event) *CloudEventMessageData {
	return &CloudEventMessageData{
		Operation:   m.formatReason(e),
		Kind:        e.Kind,
		ApiVersion:  e.ApiVersion,
		ClusterUid:  "TODO",
		Description: e.Message(),
		Obj:         e.Obj,
		OldObj:      e.OldObj,
	}
}

func (m *CloudEvent) formatReason(e event.Event) string {
	switch e.Reason {
	case "Created":
		return "create"
	case "Updated":
		return "update"
	case "Deleted":
		return "delete"
	default:
		return "unknown"
	}
}
