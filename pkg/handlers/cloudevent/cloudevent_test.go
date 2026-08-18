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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/bitnami-labs/kubewatch/pkg/filter"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCloudEventInit(t *testing.T) {
	s := &CloudEvent{}
	expectedError := fmt.Errorf(cloudEventErrMsg, "Missing cloudevent url")

	var Tests = []struct {
		cloudevent config.CloudEvent
		err        error
	}{
		{config.CloudEvent{Url: "foo"}, nil},
		{config.CloudEvent{}, expectedError},
	}

	for _, tt := range Tests {
		c := &config.Config{}
		c.Handler.CloudEvent = tt.cloudevent
		if err := s.Init(c); !reflect.DeepEqual(err, tt.err) {
			t.Fatalf("Init(): %v", err)
		}
	}
}

// sentinel is the secret material the disclosure tests look for on the wire. It
// must never appear in an outbound CloudEvent, in any encoding.
const sentinel = "s3nt1nel-D0-N0T-D1SCL0SE"

// captureCloudEvents stands in for the CloudEvent receiver and returns the raw
// bodies it was POSTed, so assertions run against the actual wire bytes.
func captureCloudEvents(t *testing.T) (*CloudEvent, *[]string) {
	t.Helper()

	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
			return
		}
		bodies = append(bodies, string(body))
	}))
	t.Cleanup(server.Close)

	return &CloudEvent{Url: server.URL, Filter: filter.NewFilter()}, &bodies
}

func sentinelSecret(suffix string) *api_v1.Secret {
	value := sentinel + suffix
	return &api_v1.Secret{
		TypeMeta:   meta_v1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: meta_v1.ObjectMeta{Name: "creds", Namespace: "default"},
		Data: map[string][]byte{
			"password": []byte(value),
			"tls.key":  []byte(value),
		},
		StringData: map[string]string{"token": value},
	}
}

// TestHandleNeverDisclosesSecretData drives the create, update and delete paths
// and asserts the Secret's bytes never reach the receiver. The update case also
// covers oldObj, which carries the *previous* secret value.
func TestHandleNeverDisclosesSecretData(t *testing.T) {
	handler, bodies := captureCloudEvents(t)

	current := sentinelSecret("-CURRENT")
	previous := sentinelSecret("-PREVIOUS")

	events := []event.Event{
		{Kind: "secret", Name: "creds", Namespace: "default", Reason: "Created", Obj: current},
		{Kind: "secret", Name: "creds", Namespace: "default", Reason: "Updated", Obj: current, OldObj: previous},
		{Kind: "secret", Name: "creds", Namespace: "default", Reason: "Deleted", Obj: current},
	}
	for _, e := range events {
		handler.Handle(e)
	}

	if len(*bodies) != len(events) {
		t.Fatalf("receiver got %d messages, want %d", len(*bodies), len(events))
	}

	for i, body := range *bodies {
		for form, encoded := range map[string]string{
			"raw":    sentinel,
			"base64": base64.StdEncoding.EncodeToString([]byte(sentinel)),
		} {
			if strings.Contains(body, encoded) {
				t.Errorf("message %d (%s): %s secret bytes disclosed: %s", i, events[i].Reason, form, body)
			}
		}

		// The notification itself must still be useful.
		if !strings.Contains(body, `"creds"`) || !strings.Contains(body, `"default"`) {
			t.Errorf("message %d (%s): lost the metadata it is a notification about: %s", i, events[i].Reason, body)
		}
	}
}

// TestHandleRedactsUnstructuredSecret covers the `customresources` path, which
// yields unstructured objects and is not gated by the `resource.secret` flag.
func TestHandleRedactsUnstructuredSecret(t *testing.T) {
	handler, bodies := captureCloudEvents(t)

	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"kind":       "Secret",
		"apiVersion": "v1",
		"metadata":   map[string]interface{}{"name": "creds", "namespace": "default"},
		"data":       map[string]interface{}{"password": base64.StdEncoding.EncodeToString([]byte(sentinel))},
		"stringData": map[string]interface{}{"token": sentinel},
	}}

	handler.Handle(event.Event{Kind: "secrets", Name: "creds", Namespace: "default", Reason: "Created", Obj: object})

	if len(*bodies) != 1 {
		t.Fatalf("receiver got %d messages, want 1", len(*bodies))
	}
	body := (*bodies)[0]
	if strings.Contains(body, sentinel) {
		t.Errorf("raw secret bytes disclosed: %s", body)
	}
	if strings.Contains(body, base64.StdEncoding.EncodeToString([]byte(sentinel))) {
		t.Errorf("base64 secret bytes disclosed: %s", body)
	}
}

// TestHandleKeepsNonSecretObjects guards the other half of the contract:
// redaction must not cost non-Secret resources their full object body, which is
// what downstream consumers match on.
func TestHandleKeepsNonSecretObjects(t *testing.T) {
	handler, bodies := captureCloudEvents(t)

	pod := &api_v1.Pod{
		TypeMeta:   meta_v1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: meta_v1.ObjectMeta{Name: "web", Namespace: "default", Labels: map[string]string{"app": "web"}},
		Spec: api_v1.PodSpec{
			NodeName:   "node-1",
			Containers: []api_v1.Container{{Name: "app", Image: "nginx:1.27"}},
		},
		Status: api_v1.PodStatus{Phase: api_v1.PodRunning},
	}

	handler.Handle(event.Event{Kind: "pod", Name: "web", Namespace: "default", ApiVersion: "v1", Reason: "Updated", Obj: pod, OldObj: pod})

	if len(*bodies) != 1 {
		t.Fatalf("receiver got %d messages, want 1", len(*bodies))
	}

	// CloudEventMessage cannot be unmarshalled back into (Obj is an interface),
	// so assert on the wire shape a receiver actually sees.
	var message struct {
		Data struct {
			Operation string                 `json:"operation"`
			Kind      string                 `json:"kind"`
			Obj       map[string]interface{} `json:"obj"`
			OldObj    map[string]interface{} `json:"oldObj"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte((*bodies)[0]), &message); err != nil {
		t.Fatalf("unmarshalling message: %v", err)
	}
	if message.Data.Operation != "update" || message.Data.Kind != "pod" {
		t.Errorf("event metadata wrong: %+v", message.Data)
	}
	for name, object := range map[string]map[string]interface{}{"obj": message.Data.Obj, "oldObj": message.Data.OldObj} {
		if object["spec"] == nil || object["status"] == nil || object["metadata"] == nil {
			t.Errorf("%s is not a full object: %+v", name, object)
		}
	}

	// Spot-check fields from across the object rather than only its metadata.
	for _, want := range []string{`"nginx:1.27"`, `"node-1"`, `"Running"`, `"app":"web"`, `"oldObj"`} {
		if !strings.Contains((*bodies)[0], want) {
			t.Errorf("non-Secret object lost %s: %s", want, (*bodies)[0])
		}
	}
}

// TestHandleDoesNotMutateInformerCacheObject guards the shared informer cache:
// the objects handed to a handler are the cache's own, and redacting one in
// place would corrupt every other reader in the process.
func TestHandleDoesNotMutateInformerCacheObject(t *testing.T) {
	handler, _ := captureCloudEvents(t)

	cached := sentinelSecret("-CACHED")
	handler.Handle(event.Event{Kind: "secret", Name: "creds", Namespace: "default", Reason: "Updated", Obj: cached, OldObj: cached})

	if got := string(cached.Data["password"]); got != sentinel+"-CACHED" {
		t.Errorf("informer cache object was mutated: Data[password] = %q", got)
	}
	if got := cached.StringData["token"]; got != sentinel+"-CACHED" {
		t.Errorf("informer cache object was mutated: StringData[token] = %q", got)
	}
}

// TestHandleKeepsEnvelopeIntactForSecrets pins down a sharp edge: objName() names
// the resource type "Secret", so the envelope's own data.kind is the literal
// string the defensive redaction layer looks for. The envelope must come out
// whole anyway — only a Secret object's data fields are ever redacted.
func TestHandleKeepsEnvelopeIntactForSecrets(t *testing.T) {
	handler, bodies := captureCloudEvents(t)

	handler.Handle(event.Event{
		Kind: "Secret", Name: "creds", Namespace: "default", ApiVersion: "v1",
		Reason: "Updated", Obj: sentinelSecret("-CURRENT"), OldObj: sentinelSecret("-PREVIOUS"),
	})

	if len(*bodies) != 1 {
		t.Fatalf("receiver got %d messages, want 1", len(*bodies))
	}

	var message struct {
		SpecVersion string `json:"specversion"`
		Type        string `json:"type"`
		ID          string `json:"id"`
		Data        struct {
			Operation   string                 `json:"operation"`
			Kind        string                 `json:"kind"`
			ApiVersion  string                 `json:"apiVersion"`
			ClusterUid  string                 `json:"clusterUid"`
			Description string                 `json:"description"`
			Obj         map[string]interface{} `json:"obj"`
			OldObj      map[string]interface{} `json:"oldObj"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte((*bodies)[0]), &message); err != nil {
		t.Fatalf("unmarshalling message: %v", err)
	}

	if message.SpecVersion != "1.0" || message.Type != "KUBERNETES_TOPOLOGY_CHANGE" || message.ID == "" {
		t.Errorf("envelope damaged: %+v", message)
	}
	if message.Data.Operation != "update" || message.Data.Kind != "Secret" ||
		message.Data.ApiVersion != "v1" || message.Data.ClusterUid == "" ||
		!strings.Contains(message.Data.Description, "creds") {
		t.Errorf("event metadata damaged: %+v", message.Data)
	}

	// Both objects are still there, still identifiable, with only data redacted.
	for name, object := range map[string]map[string]interface{}{"obj": message.Data.Obj, "oldObj": message.Data.OldObj} {
		metadata, ok := object["metadata"].(map[string]interface{})
		if !ok || metadata["name"] != "creds" || metadata["namespace"] != "default" {
			t.Errorf("%s lost its metadata: %+v", name, object)
			continue
		}
		data, ok := object["data"].(map[string]interface{})
		if !ok || len(data) != 2 {
			t.Errorf("%s should still report its 2 data keys, got: %+v", name, object["data"])
		}
	}
}
