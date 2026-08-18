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

package redact

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// sentinel is the secret material every test looks for on the way out. It must
// never appear in a redacted object, in any encoding.
const sentinel = "s3nt1nel-D0-N0T-D1SCL0SE"

// assertNoSentinel serializes value the way a handler would and fails if the
// sentinel survived, either as raw bytes or base64-encoded (the form
// encoding/json gives a Secret's []byte data).
func assertNoSentinel(t *testing.T, what string, value interface{}) {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("%s: marshal: %v", what, err)
	}

	for form, encoded := range map[string]string{
		"raw":    sentinel,
		"base64": base64.StdEncoding.EncodeToString([]byte(sentinel)),
	} {
		if strings.Contains(string(payload), encoded) {
			t.Errorf("%s: %s sentinel disclosed in %s", what, form, payload)
		}
	}
}

func secret(name string) *api_v1.Secret {
	return &api_v1.Secret{
		TypeMeta:   meta_v1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: meta_v1.ObjectMeta{Name: name, Namespace: "kube-system"},
		Type:       api_v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			"password":          []byte(sentinel),
			".dockerconfigjson": []byte(sentinel),
			"tls.key":           []byte(sentinel),
		},
		StringData: map[string]string{"token": sentinel},
	}
}

func TestObjectRedactsTypedSecret(t *testing.T) {
	original := secret("creds")
	redacted, ok := Object(original).(*api_v1.Secret)
	if !ok {
		t.Fatalf("Object() returned %T, want *v1.Secret", Object(original))
	}

	assertNoSentinel(t, "typed secret", redacted)

	// Every key is still reported, and every value is the placeholder.
	if len(redacted.Data) != len(original.Data) {
		t.Errorf("Data has %d keys, want %d", len(redacted.Data), len(original.Data))
	}
	for key, value := range redacted.Data {
		if string(value) != Placeholder {
			t.Errorf("Data[%q] = %q, want %q", key, value, Placeholder)
		}
	}
	for key, value := range redacted.StringData {
		if value != Placeholder {
			t.Errorf("StringData[%q] = %q, want %q", key, value, Placeholder)
		}
	}

	// Metadata a notification is actually for must survive untouched.
	if redacted.Name != "creds" || redacted.Namespace != "kube-system" {
		t.Errorf("metadata lost: %+v", redacted.ObjectMeta)
	}
	if redacted.Type != api_v1.SecretTypeDockerConfigJson {
		t.Errorf("Type = %q, want %q", redacted.Type, api_v1.SecretTypeDockerConfigJson)
	}
}

// The objects handed to Object() come from a shared informer cache. Redacting
// one in place would corrupt that cache for every other reader in the process.
func TestObjectDoesNotMutateInput(t *testing.T) {
	original := secret("creds")

	Object(original)

	if got := string(original.Data["password"]); got != sentinel {
		t.Errorf("input Data was mutated: got %q, want %q", got, sentinel)
	}
	if got := original.StringData["token"]; got != sentinel {
		t.Errorf("input StringData was mutated: got %q, want %q", got, sentinel)
	}
}

func TestObjectPassesThroughNonSecrets(t *testing.T) {
	pod := &api_v1.Pod{
		TypeMeta:   meta_v1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: meta_v1.ObjectMeta{Name: "web", Namespace: "default"},
		Spec:       api_v1.PodSpec{Containers: []api_v1.Container{{Name: "app", Image: "nginx"}}},
	}

	// Non-Secret objects must keep their full body: the notification payload is
	// what downstream consumers match playbooks on.
	if got := Object(pod); got != pod {
		t.Errorf("Object() copied or altered a non-Secret object: %#v", got)
	}
}

func TestObjectHandlesNil(t *testing.T) {
	// OldObj is nil on creates and deletes.
	if got := Object(nil); got != nil {
		t.Errorf("Object(nil) = %#v, want nil", got)
	}
}

func TestObjectRedactsUnstructuredSecret(t *testing.T) {
	// The `customresources` informer produces unstructured objects and is not
	// gated by the `resource.secret` flag, so a Secret can arrive this way.
	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"kind":       "Secret",
		"apiVersion": "v1",
		"metadata":   map[string]interface{}{"name": "creds", "namespace": "default"},
		"data":       map[string]interface{}{"password": base64.StdEncoding.EncodeToString([]byte(sentinel))},
		"stringData": map[string]interface{}{"token": sentinel},
	}}

	redacted := Object(object)
	assertNoSentinel(t, "unstructured secret", redacted)

	// A redacted `data` value stays valid base64, matching the typed path, so
	// receivers that decode it do not choke.
	value, _, err := unstructured.NestedString(redacted.(*unstructured.Unstructured).Object, "data", "password")
	if err != nil {
		t.Fatalf("reading redacted data: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatalf("redacted data is not valid base64: %v", err)
	}
	if string(decoded) != Placeholder {
		t.Errorf("decoded data = %q, want %q", decoded, Placeholder)
	}

	if got, _, _ := unstructured.NestedString(object.Object, "stringData", "token"); got != sentinel {
		t.Errorf("input was mutated: stringData.token = %q", got)
	}
}

func TestObjectPassesThroughUnstructuredNonSecret(t *testing.T) {
	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"kind":       "Prometheus",
		"apiVersion": "monitoring.coreos.com/v1",
		"spec":       map[string]interface{}{"replicas": int64(2)},
	}}

	if got := Object(object); got != object {
		t.Errorf("Object() altered an unstructured non-Secret: %#v", got)
	}
}

func TestJSONRedactsNestedAndListedSecrets(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte(sentinel))
	payload := []byte(`{
		"data": {
			"obj":    {"kind":"Secret","data":{"password":"` + encoded + `"},"stringData":{"token":"` + sentinel + `"}},
			"oldObj": {"kind":"Secret","data":{"password":"` + encoded + `"}},
			"items":  [{"kind":"Secret","data":{"tls.key":"` + encoded + `"}}],
			"deep":   {"wrapper":{"kind":"Secret","stringData":{"token":"` + sentinel + `"}}}
		}
	}`)

	redacted, err := JSON(payload)
	if err != nil {
		t.Fatalf("JSON(): %v", err)
	}

	if strings.Contains(string(redacted), sentinel) {
		t.Errorf("raw sentinel disclosed in %s", redacted)
	}
	if strings.Contains(string(redacted), encoded) {
		t.Errorf("base64 sentinel disclosed in %s", redacted)
	}
}

// A Secret's data field is not always a map — a hand-built or malformed payload
// can put anything there. Whatever we cannot account for is treated as secret.
func TestJSONRedactsNonMapDataField(t *testing.T) {
	payload := []byte(`{"kind":"Secret","data":"` + sentinel + `","stringData":["` + sentinel + `"]}`)

	redacted, err := JSON(payload)
	if err != nil {
		t.Fatalf("JSON(): %v", err)
	}
	if strings.Contains(string(redacted), sentinel) {
		t.Errorf("sentinel disclosed in %s", redacted)
	}
}

func TestJSONLeavesNonSecretPayloadsByteIdentical(t *testing.T) {
	// Numbers must not be pushed through float64, and a payload with nothing to
	// redact must come back exactly as it went in.
	payload := []byte(`{"kind":"Pod","metadata":{"generation":9007199254740993},"data":{"note":"kept"}}`)

	redacted, err := JSON(payload)
	if err != nil {
		t.Fatalf("JSON(): %v", err)
	}
	if string(redacted) != string(payload) {
		t.Errorf("JSON() rewrote a payload with no Secret in it:\n got %s\nwant %s", redacted, payload)
	}
}

// Redaction re-marshals the document, so numbers still have to survive it.
func TestJSONPreservesNumbersWhileRedacting(t *testing.T) {
	payload := []byte(`{"kind":"Secret","metadata":{"generation":9007199254740993},"data":{"password":"` +
		base64.StdEncoding.EncodeToString([]byte(sentinel)) + `"}}`)

	redacted, err := JSON(payload)
	if err != nil {
		t.Fatalf("JSON(): %v", err)
	}
	if !strings.Contains(string(redacted), `9007199254740993`) {
		t.Errorf("large integer lost precision: %s", redacted)
	}
	if strings.Contains(string(redacted), sentinel) {
		t.Errorf("sentinel disclosed in %s", redacted)
	}
}

func TestJSONRejectsInvalidPayload(t *testing.T) {
	// Callers must not fall back to the unredacted bytes, so an unparseable
	// payload is an error rather than a pass-through.
	if _, err := JSON([]byte(`{"kind":`)); err == nil {
		t.Error("JSON() accepted invalid JSON, want error")
	}
}
