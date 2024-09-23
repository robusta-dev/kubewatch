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

package event

import (
	"fmt"
	"os"
	"k8s.io/apimachinery/pkg/runtime"
)

// Event represent an event got from k8s api server
// Events from different endpoints need to be casted to KubewatchEvent
// before being able to be handled by handler
type Event struct {
	Namespace  string
	Kind       string
	ApiVersion string
	Component  string
	Host       string
	Reason     string
	Status     string
	Name       string
	Obj        runtime.Object
	OldObj     runtime.Object
}

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
}

// Message returns event message in standard format.
// included as a part of event packege to enhance code resuablity across handlers.
func (e *Event) Message() (msg string) {
	// using switch over if..else, since the format could vary based on the kind of the object in future.
	kubewatchName := os.Getenv("KUBEWATCH_NAME")
	switch e.Kind {
	case "namespace":
		msg = fmt.Sprintf(
			"`%s` - A namespace `%s` has been `%s`",
			kubewatchName,
			e.Name,
			e.Reason,
		)
	case "node":
		msg = fmt.Sprintf(
			"`%s` - A node `%s` has been `%s`",
			kubewatchName,
			e.Name,
			e.Reason,
		)
	case "cluster role":
		msg = fmt.Sprintf(
			"`%s` - A cluster role `%s` has been `%s`",
			kubewatchName,
			e.Name,
			e.Reason,
		)
	case "NodeReady":
		msg = fmt.Sprintf(
			"`%s` - Node `%s` is Ready : \nNodeReady",
			kubewatchName,
			e.Name,
		)
	case "NodeNotReady":
		msg = fmt.Sprintf(
			"`%s` - Node `%s` is Not Ready : \nNodeNotReady",
			kubewatchName,
			e.Name,
		)
	case "NodeRebooted":
		msg = fmt.Sprintf(
			"`%s` - Node `%s` Rebooted : \nNodeRebooted",
			kubewatchName,
			e.Name,
		)
	case "Backoff":
		msg = fmt.Sprintf(
			"`%s` - Pod `%s` in `%s` Crashed : \nCrashLoopBackOff %s",
			kubewatchName,
			e.Name,
			e.Namespace,
			e.Reason,
		)
	default:
		msg = fmt.Sprintf(
			"`%s` - A `%s` in namespace `%s` has been `%s`:\n`%s`",
			kubewatchName,
			e.Kind,
			e.Namespace,
			e.Reason,
			e.Name,
		)
	}
	return msg
}
