/*
Copyright 2016 Skippbox, Ltd.

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

package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/prometheus/client_golang/prometheus"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// sentinel is the secret material the controller must never hand to a handler.
const sentinel = "s3nt1nel-D0-N0T-D1SCL0SE"

// recordingHandler stands in for a notification handler and keeps the events it
// was given, so a test can inspect exactly what the controller emitted.
type recordingHandler struct {
	mutex  sync.Mutex
	events []event.Event
}

func (h *recordingHandler) Init(*config.Config) error { return nil }

func (h *recordingHandler) Handle(e event.Event) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.events = append(h.events, e)
}

func (h *recordingHandler) recorded() []event.Event {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return append([]event.Event(nil), h.events...)
}

// waitForEvents waits for the controller's worker to drain want events.
func (h *recordingHandler) waitForEvents(t *testing.T, want int) []event.Event {
	t.Helper()

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if recorded := h.recorded(); len(recorded) >= want {
			return recorded
		}
		time.Sleep(10 * time.Millisecond)
	}

	recorded := h.recorded()
	t.Fatalf("controller emitted %d events, want %d", len(recorded), want)
	return recorded
}

func secretInformer(client kubernetes.Interface) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets("default").List(context.Background(), options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Secrets("default").Watch(context.Background(), options)
			},
		},
		&api_v1.Secret{},
		0,
		cache.Indexers{},
	)
}

// TestSecretEventsReachHandlersRedacted drives the real trigger path — Kubernetes
// API → informer → controller → handler — and asserts the Secret's bytes are
// already gone by the time any handler sees the event, on create, update and
// delete alike. The update case also covers OldObj, which carries the previous
// secret value.
func TestSecretEventsReachHandlersRedacted(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	handler := &recordingHandler{}
	metrics := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_events_total"}, []string{"resource", "type"})

	controller := newResourceController(client, handler, secretInformer(client), "secret", V1, metrics)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	if !cache.WaitForCacheSync(stop, controller.HasSynced) {
		t.Fatal("informer cache never synced")
	}

	secret := &api_v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "creds",
			Namespace: "default",
			// The create path only notifies on objects newer than the
			// controller's start time.
			CreationTimestamp: meta_v1.NewTime(time.Now().Add(time.Minute)),
		},
		Data:       map[string][]byte{"password": []byte(sentinel)},
		StringData: map[string]string{"token": sentinel},
	}

	// Each change waits for its event before the next one is made. processItem
	// re-reads the object from the informer cache, so a burst of changes lets a
	// later one race an earlier one's notification out of existence — a
	// pre-existing kubewatch behaviour, and not what this test is about.
	created, err := client.CoreV1().Secrets("default").Create(context.Background(), secret, meta_v1.CreateOptions{})
	if err != nil {
		t.Fatalf("creating secret: %v", err)
	}
	handler.waitForEvents(t, 1)

	updated := created.DeepCopy()
	updated.Data["password"] = []byte(sentinel + "-ROTATED")
	if _, err := client.CoreV1().Secrets("default").Update(context.Background(), updated, meta_v1.UpdateOptions{}); err != nil {
		t.Fatalf("updating secret: %v", err)
	}
	handler.waitForEvents(t, 2)

	if err := client.CoreV1().Secrets("default").Delete(context.Background(), "creds", meta_v1.DeleteOptions{}); err != nil {
		t.Fatalf("deleting secret: %v", err)
	}
	events := handler.waitForEvents(t, 3)

	sawReason := map[string]bool{}
	for _, e := range events {
		sawReason[e.Reason] = true

		// Serialize the event the way a handler would and look for the bytes.
		payload, err := json.Marshal(map[string]interface{}{"obj": e.Obj, "oldObj": e.OldObj})
		if err != nil {
			t.Fatalf("marshalling %s event: %v", e.Reason, err)
		}
		for form, encoded := range map[string]string{
			"raw":    sentinel,
			"base64": base64.StdEncoding.EncodeToString([]byte(sentinel)),
		} {
			if strings.Contains(string(payload), encoded) {
				t.Errorf("%s event: %s secret bytes reached the handler: %s", e.Reason, form, payload)
			}
		}

		if e.Name != "creds" {
			t.Errorf("%s event: Name = %q, want %q", e.Reason, e.Name, "creds")
		}
	}

	for _, reason := range []string{"Created", "Updated", "Deleted"} {
		if !sawReason[reason] {
			t.Errorf("no %s event was emitted; got %v", reason, sawReason)
		}
	}
}

// TestSecretRedactionLeavesInformerCacheIntact guards the shared informer cache.
// Redacting an object in place would corrupt the cache for every other reader in
// the process, so the controller has to copy before it redacts.
func TestSecretRedactionLeavesInformerCacheIntact(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	handler := &recordingHandler{}
	metrics := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cache_test_events_total"}, []string{"resource", "type"})

	informer := secretInformer(client)
	controller := newResourceController(client, handler, informer, "secret", V1, metrics)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	if !cache.WaitForCacheSync(stop, controller.HasSynced) {
		t.Fatal("informer cache never synced")
	}

	_, err := client.CoreV1().Secrets("default").Create(context.Background(), &api_v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:              "creds",
			Namespace:         "default",
			CreationTimestamp: meta_v1.NewTime(time.Now().Add(time.Minute)),
		},
		Data: map[string][]byte{"password": []byte(sentinel)},
	}, meta_v1.CreateOptions{})
	if err != nil {
		t.Fatalf("creating secret: %v", err)
	}

	handler.waitForEvents(t, 1)

	cached, exists, err := informer.GetIndexer().GetByKey("default/creds")
	if err != nil || !exists {
		t.Fatalf("secret not in informer cache (exists=%v): %v", exists, err)
	}
	if got := string(cached.(*api_v1.Secret).Data["password"]); got != sentinel {
		t.Errorf("informer cache was mutated: Data[password] = %q, want %q", got, sentinel)
	}
}
