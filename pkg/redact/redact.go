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

// Package redact strips secret material out of Kubernetes objects before they
// leave the process in a notification.
//
// Handlers such as the CloudEvent handler serialize whole runtime.Objects into
// their payload. When Secret watching is enabled that means every Secret create,
// update and delete would ship the Secret's `data` and `stringData` — cluster,
// cloud, registry, TLS and application credentials — to an off-cluster receiver.
// Redaction happens here, once, so no handler has to remember to do it.
//
// Two layers are provided, and both are used:
//
//   - Object() is the typed layer. It runs on every event the controller emits,
//     so it protects all handlers, not just the ones that serialize objects today.
//   - JSON() is the defensive layer. It runs on the marshalled bytes immediately
//     before they are written to the wire, and catches Secrets that the typed
//     layer could not recognise — nested Secrets, and the unstructured Secrets
//     produced by the `customresources` informer, which is not gated by the
//     `resource.secret` flag.
package redact

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Placeholder replaces every redacted secret value. Key names are kept: they
// carry useful signal (which keys exist, which were added or removed) and are
// not themselves secret material, whereas the values always are.
const Placeholder = "[redacted by kubewatch]"

// secretKind is the Kubernetes kind whose data fields are always redacted.
const secretKind = "Secret"

// base64Placeholder is Placeholder as it appears on the wire under a Secret's
// `data`, whose values are []byte and so are base64-encoded by encoding/json.
// Redacting the unstructured form to the same bytes the typed form produces
// keeps a redacted `data` value decodable by receivers that expect base64.
var base64Placeholder = base64.StdEncoding.EncodeToString([]byte(Placeholder))

// dataFields are the fields holding secret material on a Secret object, mapped
// to the value each is redacted to on the wire.
var dataFields = map[string]string{
	"data":       base64Placeholder,
	"stringData": Placeholder,
}

// Object returns a copy of obj with any secret material replaced by Placeholder.
//
// The input is never mutated: objects handed to us come from a shared informer
// cache, and redacting one in place would corrupt the cache for every other
// reader. Objects that hold no secret material are returned as-is, so the
// common path costs nothing but a type check.
func Object(obj runtime.Object) runtime.Object {
	switch typed := obj.(type) {
	case nil:
		return nil
	case *api_v1.Secret:
		redacted := typed.DeepCopy()
		for key := range redacted.Data {
			redacted.Data[key] = []byte(Placeholder)
		}
		for key := range redacted.StringData {
			redacted.StringData[key] = Placeholder
		}
		return redacted
	case *unstructured.Unstructured:
		if typed.GetKind() != secretKind {
			return obj
		}
		redacted := typed.DeepCopy()
		redactUnstructuredData(redacted.Object)
		return redacted
	default:
		return obj
	}
}

// JSON redacts secret material in already-marshalled JSON, whatever shape it
// arrived in. Any object anywhere in the document whose "kind" is "Secret" has
// its "data" and "stringData" fields redacted. This is the backstop for Secrets
// the typed layer cannot see, so it deliberately keys off the wire
// representation rather than off a Go type.
//
// The document is returned unchanged if it holds no Secret, and — so that a
// redaction bug can never turn into a leak — an error is returned rather than
// the original bytes if anything about the round-trip fails.
func JSON(payload []byte) ([]byte, error) {
	// UseNumber keeps numeric fields byte-identical across the round-trip
	// instead of pushing them through float64.
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()

	var document interface{}
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}

	if !redactValue(document) {
		return payload, nil
	}

	return json.Marshal(document)
}

// redactValue walks an unmarshalled JSON value and redacts the data fields of
// every Secret it finds. It reports whether anything was redacted.
func redactValue(value interface{}) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		redacted := false
		if kind, ok := typed["kind"].(string); ok && kind == secretKind {
			redacted = redactUnstructuredData(typed)
		}
		for _, child := range typed {
			if redactValue(child) {
				redacted = true
			}
		}
		return redacted
	case []interface{}:
		redacted := false
		for _, child := range typed {
			if redactValue(child) {
				redacted = true
			}
		}
		return redacted
	default:
		return false
	}
}

// redactUnstructuredData replaces the values of a Secret's data fields in an
// unstructured object. It reports whether anything was redacted.
//
// A data field that is present but not a map is replaced wholesale: we cannot
// tell what it holds, and anything we cannot account for is treated as secret.
func redactUnstructuredData(object map[string]interface{}) bool {
	redacted := false
	for field, placeholder := range dataFields {
		value, present := object[field]
		if !present || value == nil {
			continue
		}
		entries, ok := value.(map[string]interface{})
		if !ok {
			object[field] = placeholder
			redacted = true
			continue
		}
		for key := range entries {
			entries[key] = placeholder
			redacted = true
		}
	}
	return redacted
}
